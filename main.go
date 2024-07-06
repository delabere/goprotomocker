package main

import (
	"bytes"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"os"
	"strings"
)

// extractRequestStructFromLineRange searches for a "Request" struct initialization that spans a given line number.
func extractRequestStructFromLineRange(fset *token.FileSet, node ast.Node, line int) ast.Expr {
	var structExpr ast.Expr
	ast.Inspect(node, func(n ast.Node) bool {
		if n == nil {
			return true
		}
		// Get the position information for the start and end of the current node
		startPos := fset.Position(n.Pos())
		endPos := fset.Position(n.End())
		if startPos.Line <= line && endPos.Line >= line { // Check if the given line is within the span of this node
			if cl, ok := n.(*ast.CompositeLit); ok {
				if checkRequestStruct(cl) {
					structExpr = cl
					return false // Found the struct, stop inspection
				}
			}
		}
		return true // Continue inspection to find the struct
	})
	return structExpr
}

// checkRequestStruct checks if the node is a Request struct
func checkRequestStruct(n *ast.CompositeLit) bool {
	if ident, ok := n.Type.(*ast.Ident); ok && strings.Contains(ident.Name, "Request") {
		return true
	}
	if sel, ok := n.Type.(*ast.SelectorExpr); ok {
		// This handles cases where the struct is referred with a package alias
		return strings.Contains(sel.Sel.Name, "Request")
	}
	return false
}

func main() {
	src := `package main

import "ledgerproto"

func main() {
    rsp, err := ledgerproto.CalculateBalanceRequest{
        BalanceName: ledgerproto.BalanceNameInterestPayable,
        AccountId:   pot.AccountId,
        LegalEntity: currencyLegalEntityMap[pot.Currency],
        Currency:    pot.Currency,
    }.Send(ctx).DecodeResponse()
}`

	// Parse the source code to get the AST
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "", src, parser.ParseComments)
	if err != nil {
		panic(err)
	}

	// Specify the known line number that the struct spans
	for lineNumber := 1; lineNumber <= 10; lineNumber++ {
		requestStruct := extractRequestStructFromLineRange(fset, file, lineNumber)
		if requestStruct != nil {
			var buf bytes.Buffer
			printer.Fprint(&buf, fset, requestStruct)
			os.Stdout.Write(buf.Bytes())
			break
		} else {
			println("No Request struct found spanning line", lineNumber)
		}
	}
}
