// Package analyzer contains the analyzer with the business logic of this linter.
package analyzer

import (
	"go/ast"
	"go/token"
	"go/types"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"

	"github.com/manuelarte/structinit/internal"
)

func New() *analysis.Analyzer {
	l := structinit{}

	a := &analysis.Analyzer{
		Name: "structinit",
		Doc:  "Check that structs are initialized in the desired order",
		URL:  "https://github.com/manuelarte/structinit",
		Run:  l.run,
		FactTypes: []analysis.Fact{
			&internal.HasFieldOrder{},
		},
		Requires: []*analysis.Analyzer{inspect.Analyzer},
	}

	return a
}

type (
	structinit struct{}
)

func (l *structinit) run(pass *analysis.Pass) (any, error) {
	exportStructFact(pass)
	lintStructsInitialization(pass)
	//nolint:nilnil // run api
	return nil, nil
}

// exportStructFact exports the structinit fact for the structs in the package.
func exportStructFact(pass *analysis.Pass) {
	insp, found := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	if !found {
		return
	}

	nodeFilter := []ast.Node{
		// we neext GenDecl because we filter what struct this linter applies to
		// based on in the comments there is //go:structinit
		&ast.GenDecl{},
	}

	insp.Preorder(nodeFilter, func(n ast.Node) {
		gd, isGenDecl := n.(*ast.GenDecl)
		if !isGenDecl || gd.Tok != token.TYPE {
			return
		}

		for _, spec := range gd.Specs {
			ts, isTypeSpec := spec.(*ast.TypeSpec)
			if !isTypeSpec {
				continue
			}

			doc := ts.Doc
			if doc == nil {
				doc = gd.Doc
			}

			sd, ok := internal.NewStructDecl(ts, doc)
			if !ok {
				continue
			}

			ti, hasTypeInfo := pass.TypesInfo.Defs[ts.Name]
			if !hasTypeInfo {
				continue
			}

			pass.ExportObjectFact(ti, internal.NewHasFieldOrder(sd.FieldOrder()))
		}
	})
}

func lintStructsInitialization(pass *analysis.Pass) {
	insp, found := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	if !found {
		return
	}

	nodeFilter := []ast.Node{
		&ast.CompositeLit{},
	}

	insp.Preorder(nodeFilter, func(n ast.Node) {
		cl, isCompositeLiteral := n.(*ast.CompositeLit)
		if !isCompositeLiteral {
			return
		}

		typeObj, ok := compositeLiteralType(pass, cl)
		if !ok {
			return
		}

		var fact internal.HasFieldOrder
		if !pass.ImportObjectFact(typeObj, &fact) {
			return
		}

		structInit, ok := internal.NewStructInit(cl)
		if !ok {
			return
		}

		structInit.Diagnose(pass, fact.FieldOrder())
	})
}

func compositeLiteralType(pass *analysis.Pass, cl *ast.CompositeLit) (types.Object, bool) {
	if pass.TypesInfo == nil {
		return nil, false
	}

	typ := pass.TypesInfo.TypeOf(cl)
	if typ == nil {
		return nil, false
	}

	named, isNamed := typ.(*types.Named)
	if !isNamed {
		return nil, false
	}

	obj := named.Obj()
	if obj == nil {
		return nil, false
	}

	return obj, true
}
