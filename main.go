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
		cl, ok := n.(*ast.CompositeLit)
		if !ok {
			return true
		}

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

	return newSrc, found
}

// removeAssignmentAndMethods removes the assignment and method chaining around the modified struct.
func removeAssignmentAndMethods(src []byte, fset *token.FileSet, file *ast.File, line int) (string, bool) {
	var newSrc string
	var found bool

	ast.Inspect(file, func(n ast.Node) bool {
		switch stmt := n.(type) {
		case *ast.AssignStmt:
			for _, rhs := range stmt.Rhs {
				if cl, ok := rhs.(*ast.CompositeLit); ok {
					startPos := fset.Position(stmt.Pos())
					endPos := fset.Position(stmt.End())
					if startPos.Line <= line && endPos.Line >= line && checkRequestStruct(cl) {
						// Remove the assignment and methods
						before := src[:startPos.Offset]
						after := src[endPos.Offset:]
						newSrc = string(before) + string(after)
						found = true
						return false
					}
				}
			}
		}
		return true
	})

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

	// First pass to replace the struct
	newSrc, found := extractAndReplaceStruct(src, fset, file, lineNumber)
	if !found {
		fmt.Println("No Request struct found or transformation failed")
		return
	}

	// Parse the modified source
	newFset := token.NewFileSet()
	newFile, err := parser.ParseFile(newFset, "", newSrc, parser.ParseComments)
	if err != nil {
		fmt.Printf("Failed to parse modified source: %s\n", err)
		return
	}

	// Second pass to remove assignment and methods
	finalSrc, found := removeAssignmentAndMethods([]byte(newSrc), newFset, newFile, lineNumber)
	if !found {
		fmt.Println("Failed to remove assignment or method chaining")
		return
	}

	fmt.Println("Modified Source Code:")
	fmt.Println(finalSrc)

	// Optionally, write back the modified source to the file
	if err := os.WriteFile(filePath, []byte(finalSrc), 0644); err != nil {
		fmt.Printf("Failed to write modified source back to file: %s\n", err)
		return
	}
	fmt.Println("File modified successfully.")
}
