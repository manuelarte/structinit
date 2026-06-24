package internal

import (
	"go/ast"
	"sync"
)

// StructDecl represents a struct declaration.
type StructDecl struct {
	name string
	node ast.StructType

	once       sync.Once
	fieldOrder []string
}

// NewStructDecl creates a new StructDecl. It needs to have a directive:
// //go:structinit.
func NewStructDecl(ts *ast.TypeSpec, doc *ast.CommentGroup) (StructDecl, bool) {
	st, isSt := ts.Type.(*ast.StructType)
	if !isSt {
		return StructDecl{}, false
	}

	if !containsDirective(doc) {
		return StructDecl{}, false
	}

	return StructDecl{
		name: ts.Name.Name,
		node: *st,
	}, true
}

func (s *StructDecl) FieldOrder() []string {
	s.once.Do(func() {
		if s.node.Fields == nil {
			return
		}

		for _, field := range s.node.Fields.List {
			if field.Names == nil {
				continue
			}

			for _, name := range field.Names {
				s.fieldOrder = append(s.fieldOrder, name.Name)
			}
		}
	})

	return s.fieldOrder
}
