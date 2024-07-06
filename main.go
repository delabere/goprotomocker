package main

import (
	"bytes"
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"os"
	"strings"
)

// wrapRequestStruct wraps the found struct in a testing framework setup
func wrapRequestStruct(fset *token.FileSet, node ast.Node, line int) (ast.Node, bool) {
	var structNode ast.Node
	found := false
	ast.Inspect(node, func(n ast.Node) bool {
		if cl, ok := n.(*ast.CompositeLit); ok {
			startPos := fset.Position(n.Pos())
			endPos := fset.Position(n.End())
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
					found = true
					return false // Found and transformed the struct
				}
			}
		}
		return true
	})
	return structNode, found
}

// checkRequestStruct checks if the node is a Request struct
func checkRequestStruct(n *ast.CompositeLit) bool {
	if ident, ok := n.Type.(*ast.Ident); ok && strings.Contains(ident.Name, "Request") {
		return true
	}
	if sel, ok := n.Type.(*ast.SelectorExpr); ok {
		return strings.Contains(sel.Sel.Name, "Request")
	}
	return false
}

func main() {
	// Parse command-line arguments
	var filePath string
	var lineNumber int
	flag.StringVar(&filePath, "file", "", "Path to the Go source file")
	flag.IntVar(&lineNumber, "line", 0, "Line number within the source file")
	flag.Parse()

	if filePath == "" || lineNumber == 0 {
		fmt.Println("Please specify both a file path and a line number.")
		return
	}

	// Read the source code from file
	src, err := os.ReadFile(filePath)
	if err != nil {
		fmt.Printf("Failed to read file: %s\n", err)
		return
	}

	// Parse the source code to get the AST
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, filePath, src, parser.ParseComments)
	if err != nil {
		fmt.Printf("Failed to parse file: %s\n", err)
		return
	}

	// Modify the AST based on the specified line
	modified, found := wrapRequestStruct(fset, file, lineNumber)
	if !found {
		fmt.Println("No Request struct found or transformation failed")
		return
	}

	// Print the modified AST for review
	fmt.Println("Modified AST:")
	printer.Fprint(os.Stdout, fset, modified)

	var buf bytes.Buffer
	printer.Fprint(&buf, fset, file)
	fmt.Println(buf.String())

	// // Optionally, write back the modified AST to the file
	// // Uncomment the following lines to enable writing
	// if err := os.WriteFile(filePath, buf.Bytes(), 0644); err != nil {
	// 	fmt.Printf("Failed to write modified source back to file: %s\n", err)
	// 	return
	// }
	// fmt.Println("File modified successfully.")
}
