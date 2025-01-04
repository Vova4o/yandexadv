package main

import (
    "golang.org/x/tools/go/analysis/multichecker"
    "golang.org/x/tools/go/analysis/passes/inspect"
    "golang.org/x/tools/go/analysis/passes/printf"
    "golang.org/x/tools/go/analysis/passes/shadow"
    "golang.org/x/tools/go/analysis/passes/structtag"
    "golang.org/x/tools/go/analysis/passes/unusedresult"
    "honnef.co/go/tools/staticcheck"
    "golang.org/x/tools/go/analysis"
)

func main() {
    var mychecks []*analysis.Analyzer

    // Добавление стандартных анализаторов
    mychecks = append(mychecks,
        inspect.Analyzer,
        printf.Analyzer,
        shadow.Analyzer,
        structtag.Analyzer,
        unusedresult.Analyzer,
    )

    // Добавление анализаторов из пакета staticcheck.io
    for _, v := range staticcheck.Analyzers {
        mychecks = append(mychecks, v.Analyzer)
    }

    // Добавление собственного анализатора
    mychecks = append(mychecks, noOsExitAnalyzer)

    multichecker.Main(
        mychecks...,
    )
}