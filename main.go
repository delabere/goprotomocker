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

// wrapRequestStruct wraps the found struct in a testing framework setup
func wrapRequestStruct(fset *token.FileSet, node ast.Node, line int) (ast.Node, error) {
	var structNode ast.Node
	ast.Inspect(node, func(n ast.Node) bool {
		if cl, ok := n.(*ast.CompositeLit); ok {
			startPos := fset.Position(n.Pos())
			endPos := fset.Position(n.End())
			// Check if the given line is within the start and end lines of this CompositeLit
			if startPos.Line <= line && endPos.Line >= line {
				if checkRequestStruct(cl) {
					responseType := strings.Replace(cl.Type.(*ast.SelectorExpr).Sel.Name, "Request", "Response", 1)
					responseExpr := &ast.CompositeLit{
						Type: &ast.SelectorExpr{
							X:   cl.Type.(*ast.SelectorExpr).X,
							Sel: ast.NewIdent(responseType),
						},
					}
					callExpr := &ast.CallExpr{
						Fun: &ast.SelectorExpr{
							X:   ast.NewIdent("m"),
							Sel: ast.NewIdent("ExpectRequest"),
						},
						Args: []ast.Expr{
							&ast.CallExpr{
								Fun: &ast.SelectorExpr{
									X:   ast.NewIdent("test"),
									Sel: ast.NewIdent("RequestEqualTo"),
								},
								Args: []ast.Expr{cl},
							},
						},
					}
					wrappedExpr := &ast.CallExpr{
						Fun: &ast.SelectorExpr{
							X:   callExpr,
							Sel: ast.NewIdent("RespondWith"),
						},
						Args: []ast.Expr{responseExpr},
					}
					structNode = wrappedExpr
					return false // Found and transformed the struct
				}
			}
		}
		return true
	})
	return structNode, nil
}

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
    rsp, err := foo.BarRequest{
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

	// Wrap the struct found at a specific line within its span
	lineNumber := 10 // Update as needed
	wrappedNode, err := wrapRequestStruct(fset, file, lineNumber)
	if err != nil {
		panic(err)
	}

	// Print the modified code
	if wrappedNode != nil {
		var buf bytes.Buffer
		printer.Fprint(&buf, fset, wrappedNode)
		os.Stdout.Write(buf.Bytes())
	} else {
		println("No Request struct found or transformation failed")
	}
}
