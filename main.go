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

// extractAndReplaceStruct finds the original struct and replaces it in the source code.
func extractAndReplaceStruct(src []byte, fset *token.FileSet, file *ast.File, line int) (string, bool) {
	var newSrc string
	var found bool

	ast.Inspect(file, func(n ast.Node) bool {
		switch stmt := n.(type) {
		case *ast.AssignStmt:
			fmt.Println("AssignStmt found")
			for _, rhs := range stmt.Rhs {
				// Traverse within the RHS to find the CompositeLit
				ast.Inspect(rhs, func(n ast.Node) bool {
					if cl, ok := n.(*ast.CompositeLit); ok {
						fmt.Println("CL found within AssignStmt")
						startPos := fset.Position(stmt.Pos())
						endPos := fset.Position(stmt.End())
						if startPos.Line <= line && endPos.Line >= line && checkRequestStruct(cl) {
							// Generate the new wrapped expression as a string
							wrappedExprStr := generateWrappedExpression(cl)
							// Replace the entire AssignStmt with the new expression in the source
							before := src[:startPos.Offset]
							after := src[endPos.Offset:]
							newSrc = string(before) + wrappedExprStr + string(after)
							found = true
							return false
						}
					}
					return true
				})
			}
		}
		return true
	})
	// If no AssignStmt was found, fallback to CompositeLit replacement
	if !found {
		ast.Inspect(file, func(n ast.Node) bool {
			cl, ok := n.(*ast.CompositeLit)
			if !ok {
				return true
			}

			fmt.Println("CompositeLit found")

			startPos := fset.Position(cl.Pos())
			endPos := fset.Position(cl.End())
			if startPos.Line <= line && endPos.Line >= line && checkRequestStruct(cl) {
				// Generate the new wrapped expression as a string
				wrappedExprStr := generateWrappedExpression(cl)
				// Replace the old struct text with the new expression in the source
				before := src[:startPos.Offset]
				after := src[endPos.Offset:]
				newSrc = string(before) + wrappedExprStr + string(after)
				found = true
				return false
			}
			return true
		})
	}

	return newSrc, found
}

// generateWrappedExpression generates the modified expression as a string.
func generateWrappedExpression(cl *ast.CompositeLit) string {
	var builder strings.Builder
	responseType := strings.Replace(cl.Type.(*ast.SelectorExpr).Sel.Name, "Request", "Response", 1)

	// Begin the wrapped expression
	fmt.Fprintf(&builder, "m.ExpectRequest(test.RequestEqualTo(%s{\n", cl.Type.(*ast.SelectorExpr).Sel.Name)

	// Iterate through the fields in the CompositeLit and add them to the expression
	for _, elt := range cl.Elts {
		if kv, ok := elt.(*ast.KeyValueExpr); ok {
			key := kv.Key.(*ast.Ident).Name
			valBuf := &bytes.Buffer{}
			printer.Fprint(valBuf, token.NewFileSet(), kv.Value)
			val := valBuf.String()
			fmt.Fprintf(&builder, "\t%s: %s,\n", key, val)
		}
	}

	// Close the original struct and add the response struct
	fmt.Fprintf(&builder, "})).RespondWith(%s{})", responseType)

	return builder.String()
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
	newSrc, found := extractAndReplaceStruct(src, fset, file, lineNumber)
	if !found {
		fmt.Println("No Request struct found or transformation failed")
		return
	}

	// // Optionally, write back the modified source to the file
	// if err := os.WriteFile(filePath, []byte(newSrc), 0644); err != nil {
	// 	fmt.Printf("Failed to write modified source back to file: %s\n", err)
	// 	return
	// }
	// fmt.Println("File modified successfully.")
	fmt.Println("Modified Source Code:")
	fmt.Println(newSrc)
}
