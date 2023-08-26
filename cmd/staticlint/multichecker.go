package main

import (
	"strings"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"
	"golang.org/x/tools/go/analysis/passes/assign"
	"golang.org/x/tools/go/analysis/passes/atomic"
	"golang.org/x/tools/go/analysis/passes/copylock"
	"golang.org/x/tools/go/analysis/passes/httpresponse"
	"golang.org/x/tools/go/analysis/passes/lostcancel"
	"golang.org/x/tools/go/analysis/passes/printf"
	"golang.org/x/tools/go/analysis/passes/shadow"
	"golang.org/x/tools/go/analysis/passes/shift"
	"golang.org/x/tools/go/analysis/passes/structtag"
	"golang.org/x/tools/go/analysis/passes/tests"
	"golang.org/x/tools/go/analysis/passes/unmarshal"
	"honnef.co/go/tools/staticcheck"
)

func main() {
	var checks []*analysis.Analyzer
	checks = append(checks, []*analysis.Analyzer{
		assign.Analyzer,
		atomic.Analyzer,
		copylock.Analyzer,
		httpresponse.Analyzer,
		lostcancel.Analyzer,
		printf.Analyzer,
		shadow.Analyzer,
		shift.Analyzer,
		tests.Analyzer,
		unmarshal.Analyzer,
		structtag.Analyzer,
		ExitCheckAnalyzer,
	}...)

	for _, v := range staticcheck.Analyzers {
		if strings.HasPrefix(v.Analyzer.Name, "SA") || strings.HasPrefix(v.Analyzer.Name, "S1") {
			checks = append(checks, v.Analyzer)
		}
	}
	multichecker.Main(
		checks...,
	)
}
