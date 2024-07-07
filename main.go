package main

import (
	"bytes"
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"strings"

	"github.com/dave/dst"
	"github.com/dave/dst/decorator"
	"github.com/dave/dst/dstutil"
)

func extractAndReplaceAst(fset *token.FileSet, file *dst.File, dec *decorator.Decorator, line int) {
	dstutil.Apply(file, func(cr *dstutil.Cursor) bool {
		switch n := cr.Node().(type) {
		case *dst.AssignStmt:
			for _, rhs := range n.Rhs {
				dst.Inspect(rhs, func(n dst.Node) bool {
					if cl, ok := n.(*dst.CompositeLit); ok {
						orig := dec.Ast.Nodes[cl].(*ast.CompositeLit)
						startPos := fset.Position(orig.Pos())
						endPos := fset.Position(orig.End())
						if startPos.Line <= line && endPos.Line >= line {
							wrappedExpr := generateWrappedExpressionAsDst(cl)
							cr.Replace(&dst.ExprStmt{X: wrappedExpr})
							return false
						}
						return true
					}
					return true
				})
			}
		case *dst.CompositeLit:
			orig := dec.Ast.Nodes[n].(*ast.CompositeLit)
			startPos := fset.Position(orig.Pos())
			endPos := fset.Position(orig.End())
			if startPos.Line <= line && endPos.Line >= line {
				wrappedExpr := generateWrappedExpressionAsDst(n)
				cr.Replace(wrappedExpr)
				return false
			}
			return true
		}
		return true
	}, nil)
}

func generateWrappedExpressionAsDst(cl *dst.CompositeLit) *dst.CallExpr {
	ok, s := checkRequestStruct(cl)
	if !ok {
		return nil
	}

	switch s {
	case "request":
		// Create a deep copy of cl to use in the new expression
		clCopy := cloneCompositeLit(cl)

		callExpr := &dst.CallExpr{
			Fun: &dst.SelectorExpr{
				X:   dst.NewIdent("m"),
				Sel: dst.NewIdent("ExpectRequest"),
			},
			Args: []dst.Expr{
				&dst.CallExpr{
					Fun: &dst.SelectorExpr{
						X:   dst.NewIdent("test"),
						Sel: dst.NewIdent("RequestEqualTo"),
					},
					Args: []dst.Expr{clCopy}, // Use the copied composite literal here
				},
			},
		}

		responseType := strings.Replace(cl.Type.(*dst.SelectorExpr).Sel.Name, "Request", "Response", 1)
		responseExpr := &dst.CompositeLit{
			Type: &dst.SelectorExpr{
				X:   cl.Type.(*dst.SelectorExpr).X,
				Sel: dst.NewIdent(responseType),
			},
		}
		wrappedExpr := &dst.CallExpr{
			Fun: &dst.SelectorExpr{
				X:   callExpr,
				Sel: dst.NewIdent("RespondWith"),
			},
			Args: []dst.Expr{responseExpr},
		}

		return wrappedExpr
	case "event":

		// Create a deep copy of cl to use in the new expression
		clCopy := cloneCompositeLit(cl)
		callExpr := &dst.CallExpr{
			Fun: &dst.SelectorExpr{
				X:   dst.NewIdent("m"),
				Sel: dst.NewIdent("ExpectFirehoseEvent"),
			},
			Args: []dst.Expr{
				&dst.CallExpr{
					Fun: &dst.SelectorExpr{
						X:   dst.NewIdent("test"),
						Sel: dst.NewIdent("EventMatching"),
					},
					Args: []dst.Expr{clCopy}, // Use the copied composite literal here
				},
			},
		}

		return callExpr
	}
	
	return nil
}

// cloneCompositeLit creates a deep copy of a dst.CompositeLit
func cloneCompositeLit(orig *dst.CompositeLit) *dst.CompositeLit {
	if orig == nil {
		return nil
	}
	copy := &dst.CompositeLit{
		Elts: make([]dst.Expr, len(orig.Elts)),
	}
	for i, elt := range orig.Elts {
		copy.Elts[i] = dst.Clone(elt).(dst.Expr)
	}
	if orig.Type != nil {
		copy.Type = dst.Clone(orig.Type).(dst.Expr)
	}
	return copy
}

func checkRequestStruct(cr *dst.CompositeLit) (bool, string) {
	switch n := cr.Type.(type) {
	case *dst.Ident:
		if strings.Contains(n.Name, "Request") {
			return true, "request"
		} else if strings.Contains(n.Name, "Event") {
			return true, "event"
		}
	case *dst.SelectorExpr:
		if strings.Contains(n.Sel.Name, "Request") {
			return true, "request"
		} else if strings.Contains(n.Sel.Name, "Event") {
			return true, "event"
		}
	}
	return false, ""
}

func parseFile(filePath string, lineNumber int) (bytes.Buffer, error) {
	src, err := os.ReadFile(filePath)
	if err != nil {
		fmt.Printf("Failed to read file: %s\n", err)
		return bytes.Buffer{}, err
	}

	fset := token.NewFileSet()
	astFile, err := parser.ParseFile(fset, filePath, src, parser.ParseComments)
	if err != nil {
		fmt.Printf("Failed to parse file: %s\n", err)
		panic(err)
	}

	return parse(fset, astFile, lineNumber)
}

func parseBytes(src []byte, lineNumber int) (bytes.Buffer, error) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test-file", src, parser.ParseComments)
	if err != nil {
		return bytes.Buffer{}, err
	}

	return parse(fset, file, lineNumber)
}

func parse(fset *token.FileSet, astFile *ast.File, lineNumber int) (bytes.Buffer, error) {
	dec := decorator.NewDecorator(fset)
	file, err := dec.DecorateFile(astFile)
	if err != nil {
		panic(err)
	}

	// First pass to replace the struct or AssignStmt
	extractAndReplaceAst(fset, file, dec, lineNumber)

	// put the dst into a buffer
	var buf bytes.Buffer
	if err := decorator.Fprint(&buf, file); err != nil {
		fmt.Printf("Error printing AST: %s\n", err)
		return bytes.Buffer{}, nil
	}
	return buf, err
}

func main() {
	var filePath string
	var lineNumber int
	var write bool
	flag.StringVar(&filePath, "file", "", "Path to the Go source file")
	flag.IntVar(&lineNumber, "line", 0, "Line number within the source file")
	flag.BoolVar(&write, "write", false, "Whether to write to the source file")
	flag.Parse()

	if filePath == "" || lineNumber == 0 {
		fmt.Println("Please specify both a file path and a line number.")
		return
	}

	buf, err := parseFile(filePath, lineNumber)
	if err != nil {
		fmt.Printf("error parsing ast %s", err)
		return
	}

	if write {
		// Optionally, write back the modified source to the file
		if err := os.WriteFile(filePath, buf.Bytes(), 0644); err != nil {
			fmt.Printf("Failed to write modified source back to file: %s\n", err)
			return
		}
		fmt.Println("File modified successfully.")
		fmt.Println("Modified Source Code:")
	}
}
