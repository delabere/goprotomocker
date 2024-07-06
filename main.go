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

	"golang.org/x/tools/go/ast/astutil"
)

func extractAndReplaceAst(fset *token.FileSet, file *ast.File, line int) {
	astutil.Apply(file, func(cr *astutil.Cursor) bool {
		switch n := cr.Node().(type) {
		case *ast.AssignStmt:
			fmt.Println("AssignStmt found")
			for _, rhs := range n.Rhs {
				ast.Inspect(rhs, func(n ast.Node) bool {
					if cl, ok := n.(*ast.CompositeLit); ok {
						startPos := fset.Position(n.Pos())
						endPos := fset.Position(n.End())
						if startPos.Line <= line && endPos.Line >= line && checkRequestStruct(cl) {
							fmt.Println("CL found within AssignStmt")
							wrappedExpr := generateWrappedExpressionAsAst(cl)
							cr.Replace(&ast.ExprStmt{X: wrappedExpr})
							return false
						}
						return true
					}
					return true
				})
			}

		case *ast.CompositeLit:
			startPos := fset.Position(n.Pos())
			endPos := fset.Position(n.End())
			if startPos.Line <= line && endPos.Line >= line && checkRequestStruct(n) {
				fmt.Println("CL found within AssignStmt")
				wrappedExpr := generateWrappedExpressionAsAst(n)
				cr.Replace(wrappedExpr)
				return false
			}
			return true
		}
		return true
	}, nil)

}

// generateWrappedExpressionAsAst generates the modified expression as an ast node
func generateWrappedExpressionAsAst(cl *ast.CompositeLit) *ast.CallExpr {
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

		return wrappedExpr
	}

	return nil
}

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
	var filePath string
	var lineNumber int
	flag.StringVar(&filePath, "file", "", "Path to the Go source file")
	flag.IntVar(&lineNumber, "line", 0, "Line number within the source file")
	flag.Parse()

	if filePath == "" || lineNumber == 0 {
		fmt.Println("Please specify both a file path and a line number.")
		return
	}

	src, err := os.ReadFile(filePath)
	if err != nil {
		fmt.Printf("Failed to read file: %s\n", err)
		return
	}

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, filePath, src, parser.ParseComments)
	if err != nil {
		fmt.Printf("Failed to parse file: %s\n", err)
		return
	}

	// First pass to replace the struct or AssignStmt
	extractAndReplaceAst(fset, file, lineNumber)

	// Print the parsed AST
	var buf bytes.Buffer
	if err := printer.Fprint(&buf, fset, file); err != nil {
		fmt.Printf("Error printing AST: %s\n", err)
		return
	}
	fmt.Println("Original Source Code:")
	fmt.Println(buf.String())

	// // Optionally, write back the modified source to the file
	// if err := os.WriteFile(filePath, []byte(newSrc), 0644); err != nil {
	// 	fmt.Printf("Failed to write modified source back to file: %s\n", err)
	// 	return
	// }
	// fmt.Println("File modified successfully.")
	// fmt.Println("Modified Source Code:")
	// fmt.Println(newSrc)
}
