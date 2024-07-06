package main

import (
	"bytes"
	"go/parser"
	"go/printer"
	"go/token"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/tools/imports"
)

func TestMainExpression(t *testing.T) {
	src := `package main

func main() {
	_, err = fooproto.BarRequest{
		IdempotencyKey:    idempotencyKey,
		SubjectId:         subjectID,
		Trigger:           fooproto.SomeFooConst,
	}.Send(ctx).DecodeResponse()
	if err != nil {
		panic("uh oh")	
	}
}`

	expected := `package main

func main() {
	m.ExpectRequest(test.RequestEqualTo(fooproto.BarRequest{
		IdempotencyKey: idempotencyKey,
		SubjectId:      subjectID,
		Trigger:        fooproto.SomeFooConst,
	})).RespondWith(fooproto.BarResponse{})

	if err != nil {
		panic("uh oh")
	}
}`

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test-file", src, parser.ParseComments)
	assert.NoError(t, err)

	// First pass to replace the struct or AssignStmt
	extractAndReplaceAst(fset, file, 6)

	// Print the parsed AST
	var buf bytes.Buffer
	err = printer.Fprint(&buf, fset, file)
	assert.NoError(t, err)

	// Format the output to preserve the original formatting and whitespace
	formattedSrc, err := imports.Process("test-file", buf.Bytes(), nil)
	assert.NoError(t, err)

	// format.Node(&buf, fset, file)

	assert.Equal(t, expected, strings.Trim(string(formattedSrc), "\n"))
}

func TestMainDeclaration(t *testing.T) {
	src := `package main

func main() {
	fooproto.BarRequest{
		IdempotencyKey:    idempotencyKey,
		SubjectId:         subjectID,
		Trigger:           fooproto.SomeFooConst,
	}
}`

	expected := `package main

func main() {
	m.ExpectRequest(test.RequestEqualTo(fooproto.BarRequest{
		IdempotencyKey: idempotencyKey,
		SubjectId:      subjectID,
		Trigger:        fooproto.SomeFooConst,
	})).RespondWith(fooproto.BarResponse{})

}`

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test-file", src, parser.ParseComments)
	assert.NoError(t, err)

	// First pass to replace the struct or AssignStmt
	extractAndReplaceAst(fset, file, 6)

	// Print the parsed AST
	var buf bytes.Buffer
	err = printer.Fprint(&buf, fset, file)
	assert.NoError(t, err)

	// Format the output to preserve the original formatting and whitespace
	formattedSrc, err := imports.Process("test-file", buf.Bytes(), nil)
	assert.NoError(t, err)

	// format.Node(&buf, fset, file)

	assert.Equal(t, expected, strings.Trim(string(formattedSrc), "\n"))
}

func TestWithComment(t *testing.T) {
	src := `package main

func main() {
	// There is a comment here
	_, err = fooproto.BarRequest{
		IdempotencyKey:    idempotencyKey,
		SubjectId:         subjectID,
		Trigger:           fooproto.SomeFooConst,
	}.Send(ctx).DecodeResponse()
	if err != nil {
		panic("uh oh")	
	}
}`

	expected := `package main

func main() {
	// There is a comment here
	m.ExpectRequest(test.RequestEqualTo(fooproto.BarRequest{
		IdempotencyKey: idempotencyKey,
		SubjectId:      subjectID,
		Trigger:        fooproto.SomeFooConst,
	})).RespondWith(fooproto.BarResponse{})

	if err != nil {
		panic("uh oh")
	}
}`

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test-file", src, parser.ParseComments)
	assert.NoError(t, err)

	// First pass to replace the struct or AssignStmt
	extractAndReplaceAst(fset, file, 6)

	// Print the parsed AST
	var buf bytes.Buffer
	err = printer.Fprint(&buf, fset, file)
	assert.NoError(t, err)

	// Format the output to preserve the original formatting and whitespace
	formattedSrc, err := imports.Process("test-file", buf.Bytes(), nil)
	assert.NoError(t, err)

	assert.Equal(t, expected, strings.Trim(string(formattedSrc), "\n"))
}

func TestWithExpressionBefore(t *testing.T) {
	src := `package main

func main() {
	println("hello world")

	_, err = fooproto.BarRequest{
		IdempotencyKey:    idempotencyKey,
		SubjectId:         subjectID,
		Trigger:           fooproto.SomeFooConst,
	}.Send(ctx).DecodeResponse()
	if err != nil {
		panic("uh oh")	
	}
}`

	expected := `package main

func main() {
	println("hello world")

	m.ExpectRequest(test.RequestEqualTo(fooproto.BarRequest{
		IdempotencyKey: idempotencyKey,
		SubjectId:      subjectID,
		Trigger:        fooproto.SomeFooConst,
	})).RespondWith(fooproto.BarResponse{})

	if err != nil {
		panic("uh oh")
	}
}`

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test-file", src, parser.ParseComments)
	assert.NoError(t, err)

	// First pass to replace the struct or AssignStmt
	extractAndReplaceAst(fset, file, 6)

	// Print the parsed AST
	var buf bytes.Buffer
	err = printer.Fprint(&buf, fset, file)
	assert.NoError(t, err)

	// Format the output to preserve the original formatting and whitespace
	formattedSrc, err := imports.Process("test-file", buf.Bytes(), nil)
	assert.NoError(t, err)

	assert.Equal(t, expected, strings.Trim(string(formattedSrc), "\n"))
}
