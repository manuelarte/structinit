package internal

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/printer"
	"go/token"
	"strings"
)

func containsDirective(doc *ast.CommentGroup) bool {
	if doc == nil {
		return false
	}

	for _, comment := range doc.List {
		if strings.HasPrefix(comment.Text, "//go:structinit") {
			return true
		}
	}

	return false
}

// nodeBytes renders any ast.Node as source bytes.
func nodeBytes(fset *token.FileSet, n ast.Node) ([]byte, error) {
	var buf bytes.Buffer

	if err := printer.Fprint(&buf, fset, n); err != nil {
		return nil, fmt.Errorf("failed to render node: %w", err)
	}

	return buf.Bytes(), nil
}
