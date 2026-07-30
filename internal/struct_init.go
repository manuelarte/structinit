package internal

import (
	"go/ast"
	"go/token"
	"slices"

	"golang.org/x/tools/go/analysis"
)

type (
	// StructInit represents a struct initialization. It has to be a composite literal in which
	// the fields follow this pattern:
	// - key: *ast.Ident
	// - value: *ast.BasicLit | *ast.CompositeLit | *ast.Ident | *ast.CallExpr (when there is only one call expression)
	// any other combination is not supported.
	StructInit struct {
		compositeLiteral *ast.CompositeLit

		keyValueExprs []*keyValueExpr
	}

	keyValueExpr struct {
		n *ast.KeyValueExpr

		// the comment declared on the top of the field initialization.
		topComment *ast.CommentGroup
		// the comment declared in the same line after the field initialization.
		inlinedComment *ast.CommentGroup
		// original index where the field was declared in the composite lit.
		originalIndex int
		// expected index where the field was declared in the composite lit.
		expectedIndex int
	}
)

// NewStructInit returns a StructInit if the given composite literal is a struct initialization.
func NewStructInit(cl *ast.CompositeLit) (StructInit, bool) {
	numberOfCallExpr := 0
	keyValueExprs := make([]*keyValueExpr, 0, len(cl.Elts))
	for originalIndex, elt := range cl.Elts {
		kv, isKeyValue := elt.(*ast.KeyValueExpr)
		if !isKeyValue {
			return StructInit{}, false
		}

		switch kv.Value.(type) {
		case *ast.Ident, *ast.BasicLit, *ast.CompositeLit:
			keyValueExprs = append(keyValueExprs, &keyValueExpr{
				n:             kv,
				originalIndex: originalIndex,
				expectedIndex: -1,
			})
		case *ast.CallExpr:
			numberOfCallExpr++
			if numberOfCallExpr > 1 {
				return StructInit{}, false
			}
			keyValueExprs = append(keyValueExprs, &keyValueExpr{
				n:             kv,
				originalIndex: originalIndex,
				expectedIndex: -1,
			})

		default:
			return StructInit{}, false
		}
	}

	return StructInit{
		compositeLiteral: cl,
		keyValueExprs:    keyValueExprs,
	}, true
}

// Diagnose checks if the fields are initialized in the correct order.
func (s StructInit) Diagnose(pass *analysis.Pass, expectedOrder []string) {
	// First, we populate our helper struct keyValueExpr with the information we need to
	// check if the fields are ordered or not.
	if len(expectedOrder) == 0 {
		return
	}

	hasFieldNotOrdered := s.assignIndexes(expectedOrder)
	if !hasFieldNotOrdered {
		return
	}

	s.assignComments(pass)

	sfs, err := s.buildSuggestedFixes(pass)
	if err != nil {
		// if I can't build suggested fix, then I don't propose a suggested fix
		sfs = nil
	}

	pass.Report(analysis.Diagnostic{
		Pos:            s.compositeLiteral.Pos(),
		End:            s.compositeLiteral.End(),
		Category:       "field-order",
		Message:        "fields are not initialized in declared order",
		URL:            "https://github.com/manuelarte/structinit#feature",
		SuggestedFixes: sfs,
		Related:        s.buildRelated(),
	})
}

func (s StructInit) applicableFields(allOrderedFields []string) []string {
	return slices.DeleteFunc(allOrderedFields, func(name string) bool {
		keyedElementsNames := make([]string, 0)

		for _, elt := range s.compositeLiteral.Elts {
			kv, isKeyValue := elt.(*ast.KeyValueExpr)
			if !isKeyValue {
				continue
			}

			if kv.Key == nil {
				continue
			}

			ident, isIdent := kv.Key.(*ast.Ident)
			if !isIdent {
				continue
			}

			keyedElementsNames = append(keyedElementsNames, ident.Name)
		}

		return !slices.Contains(keyedElementsNames, name)
	})
}

func (s StructInit) assignIndexes(expectedOrder []string) bool {
	expectedOrder = s.applicableFields(expectedOrder)
	hasFieldNotOrdered := false

	for expectedIndex, name := range expectedOrder {
		if originalIndex := slices.IndexFunc(s.keyValueExprs, func(expr *keyValueExpr) bool {
			exprIdent, isIdent := expr.n.Key.(*ast.Ident)

			return isIdent && exprIdent.Name == name
		}); originalIndex != -1 {
			if originalIndex != expectedIndex {
				hasFieldNotOrdered = true
			}

			x := s.keyValueExprs[originalIndex]
			x.expectedIndex = expectedIndex
		}
	}

	return hasFieldNotOrdered
}

func (s StructInit) assignComments(pass *analysis.Pass) {
	comments, _ := s.getCommentGroups(pass)

	isInlinedComment := len(comments) >= 1 && comments[0].Pos() == s.compositeLiteral.Lbrace+2
	if isInlinedComment {
		comments = comments[1:]
	}

	for _, comment := range comments {
		for i, kv := range s.keyValueExprs {
			// Determine the boundary: previous keyValueExpr end or start of composite literal
			var previousEnd token.Pos
			if i > 0 {
				previousEnd = s.keyValueExprs[i-1].n.End()
			} else {
				previousEnd = s.compositeLiteral.Lbrace
			}

			// Check if comment is after previous keyValueExpr and before/at current one
			if comment.Pos() > previousEnd && comment.End() <= kv.n.End() {
				// Check if it's an inlined comment (same line as keyValueExpr)
				fileSet := pass.Fset
				commentLine := fileSet.Position(comment.Pos()).Line
				kvLine := fileSet.Position(kv.n.End()).Line

				if commentLine == kvLine {
					kv.inlinedComment = comment
				} else if comment.Pos() < kv.n.Pos() {
					// Top comment (before the keyValueExpr)
					kv.topComment = comment
				}

				break // Move to next comment
			}
		}
	}
}

func (s StructInit) buildRelated() []analysis.RelatedInformation {
	related := make([]analysis.RelatedInformation, 0)

	for i, currentKeyValueExpr := range s.keyValueExprs {
		previousIndex := i - 1
		if previousIndex < 0 {
			continue
		}

		previousKeyValueExpr := s.keyValueExprs[previousIndex]
		if previousKeyValueExpr.expectedIndex < currentKeyValueExpr.expectedIndex {
			continue
		}

		related = append(related, analysis.RelatedInformation{
			Pos:     previousKeyValueExpr.n.Pos(),
			End:     previousKeyValueExpr.n.End(),
			Message: "field initialized out of position",
		})
	}

	return related
}

func (s StructInit) buildSuggestedFixes(
	pass *analysis.Pass,
) ([]analysis.SuggestedFix, error) {
	textEdits := make([]analysis.TextEdit, 0)
	fset := pass.Fset

	for _, mf := range s.keyValueExprs {
		if mf.expectedIndex == -1 {
			return nil, nil
		}

		originalKv := s.keyValueExprs[mf.originalIndex]
		expectedKv := s.keyValueExprs[mf.expectedIndex]

		var newTextBytes []byte

		// Include top comment if present
		if originalKv.topComment != nil {
			topCommentBytes, err := nodeBytes(fset, originalKv.topComment)
			if err != nil {
				return nil, err
			}

			newTextBytes = append(newTextBytes, topCommentBytes...)
			newTextBytes = append(newTextBytes, '\n')
		}

		// Include the keyValueExpr itself
		kvBytes, err := nodeBytes(fset, originalKv.n)
		if err != nil {
			return nil, err
		}

		newTextBytes = append(newTextBytes, kvBytes...)

		// Include inlined comment if present
		if originalKv.inlinedComment != nil {
			newTextBytes = append(newTextBytes, ' ')

			inlinedCommentBytes, errByte := nodeBytes(fset, originalKv.inlinedComment)
			if errByte != nil {
				return nil, errByte
			}

			newTextBytes = append(newTextBytes, inlinedCommentBytes...)
		}

		textEdits = append(textEdits, analysis.TextEdit{
			Pos:     expectedKv.n.Pos(),
			End:     expectedKv.n.End(),
			NewText: newTextBytes,
		})
	}

	return []analysis.SuggestedFix{
		{
			Message:   "reorder fields to match struct declaration",
			TextEdits: textEdits,
		},
	}, nil
}

func (s StructInit) getCommentGroups(pass *analysis.Pass) ([]*ast.CommentGroup, bool) {
	var file *ast.File

	for _, f := range pass.Files {
		if f.Pos() <= s.compositeLiteral.Pos() && s.compositeLiteral.Pos() <= f.End() {
			file = f

			break
		}
	}

	if file == nil {
		return nil, false
	}

	comments := make([]*ast.CommentGroup, 0, len(file.Comments))
	for _, comment := range file.Comments {
		if comment.Pos() >= s.compositeLiteral.Pos() && comment.End() <= s.compositeLiteral.End() {
			comments = append(comments, comment)
		}
	}

	return comments, true
}
