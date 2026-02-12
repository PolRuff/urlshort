package main

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"
)

func TestForbidExitAnalyzer(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), ForbidExitAnalyzer, "./...")
}
