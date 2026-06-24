package main

import (
	"golang.org/x/tools/go/analysis/singlechecker"

	"github.com/manuelarte/structinit/analyzer"
)

func main() {
	singlechecker.Main(analyzer.New())
}
