// Package plugin add supports to adding this analyzer as a golangci-lint
// plugin.
package plugin

import (
	"github.com/golangci/plugin-module-register/register"
	"golang.org/x/tools/go/analysis"

	"github.com/manuelarte/structinit/analyzer"
)

//nolint:gochecknoinits // init needed for plugin
func init() {
	register.Plugin("structinit", New)
}

func New(_ any) (register.LinterPlugin, error) {
	return &structinitPlugin{}, nil
}

var _ register.LinterPlugin = new(structinitPlugin)

type structinitPlugin struct{}

func (u structinitPlugin) BuildAnalyzers() ([]*analysis.Analyzer, error) {
	return []*analysis.Analyzer{
		analyzer.New(),
	}, nil
}

func (u structinitPlugin) GetLoadMode() string {
	return register.LoadModeSyntax
}
