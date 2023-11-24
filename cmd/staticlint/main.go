package main

import (
	gocritic "github.com/go-critic/go-critic/checkers/analyzer"
	"github.com/timakin/bodyclose/passes/bodyclose"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"
	"golang.org/x/tools/go/analysis/passes/fieldalignment"
	"golang.org/x/tools/go/analysis/passes/printf"
	"golang.org/x/tools/go/analysis/passes/shadow"
	"golang.org/x/tools/go/analysis/passes/structtag"
	"honnef.co/go/tools/simple"
	"honnef.co/go/tools/staticcheck"
	"honnef.co/go/tools/stylecheck"

	"github.com/MowlCoder/go-url-shortener/cmd/staticlint/osexitmain"
)

func main() {
	analyzers := []*analysis.Analyzer{
		shadow.Analyzer,
		printf.Analyzer,
		structtag.Analyzer,
		osexitmain.Analyzer,
	}

	simpleChecks := map[string]bool{
		"S1012": true,
		"S1024": true,
		"S1025": true,
		"S1011": true,
	}
	for _, check := range simple.Analyzers {
		if !simpleChecks[check.Analyzer.Name] {
			continue
		}

		analyzers = append(analyzers, check.Analyzer)
	}

	styleChecks := map[string]bool{
		"ST1006": true,
		"ST1017": true,
		"ST1023": true,
		"ST1015": true,
	}
	for _, check := range stylecheck.Analyzers {
		if !styleChecks[check.Analyzer.Name] {
			continue
		}

		analyzers = append(analyzers, check.Analyzer)
	}

	for _, check := range staticcheck.Analyzers {
		analyzers = append(analyzers, check.Analyzer)
	}

	analyzers = append(
		analyzers,
		gocritic.Analyzer,
		bodyclose.Analyzer,
		fieldalignment.Analyzer,
	)

	multichecker.Main(
		analyzers...,
	)
}
