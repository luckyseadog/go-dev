package main

import (
	"strings"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"
	"golang.org/x/tools/go/analysis/passes/assign"
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
	// Analyzers array for package.
	var checks []*analysis.Analyzer
	checks = append(checks, []*analysis.Analyzer{
		assign.Analyzer,       // Analyzer that checks for useless assignments.
		copylock.Analyzer,     // Analyzer that checks for locks erroneously passed by value
		httpresponse.Analyzer, // Analyzer that checks for mistakes using HTTP responses.
		lostcancel.Analyzer,   // Analyzer that checks for failure to call a context cancellation function.
		printf.Analyzer,       // Analyzer that checks consistency of Printf format strings and arguments.
		shadow.Analyzer,       // Analyzer that checks for shadowed variables.
		shift.Analyzer,        // Analyzer that checks for shifts that exceed the width of an integer.
		tests.Analyzer,        // Analyzer that checks for common mistaken usages of tests and examples.
		unmarshal.Analyzer,    // Analyzer that checks for passing non-pointer or non-interface types to unmarshal and decode functions.
		structtag.Analyzer,    // Analyzer that checks struct field tags are well-formed.
		ExitCheckAnalyzer,     // Analyzer that checks for inappropriate use of os.Exit
	}...)

	// Add analyzers for Various misuses of the standard library, Concurrency issues,
	// Testing issues, Correctness issues, Performance issues etc.
	for _, v := range staticcheck.Analyzers {
		if strings.HasPrefix(v.Analyzer.Name, "SA") || strings.HasPrefix(v.Analyzer.Name, "S1") {
			checks = append(checks, v.Analyzer)
		}
	}

	// Finally, build multichecker with array of checks.
	multichecker.Main(
		checks...,
	)
}
