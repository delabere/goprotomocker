package main

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMainTransformations(t *testing.T) {
	tests := []struct {
		name       string
		src        string
		expected   string
		lineNumber int
	}{
		{
			name: "Expression, first line",
			src: `package main

func main() {
	rsp, err := fooproto.BarRequest{
		IdempotencyKey:    idempotencyKey,
		SubjectId:         subjectID,
		Trigger:           fooproto.SomeFooConst,
	}.Send(ctx).DecodeResponse()
}`,
			expected: `package main

func main() {
	m.ExpectRequest(test.RequestEqualTo(fooproto.BarRequest{
		IdempotencyKey: idempotencyKey,
		SubjectId:      subjectID,
		Trigger:        fooproto.SomeFooConst,
	})).RespondWith(fooproto.BarResponse{})
}`,
			lineNumber: 4,
		},
		{
			name: "Expression, last line",
			src: `package main

func main() {
	rsp, err := fooproto.BarRequest{
		IdempotencyKey:    idempotencyKey,
		SubjectId:         subjectID,
		Trigger:           fooproto.SomeFooConst,
	}.Send(ctx).DecodeResponse()
}`,
			expected: `package main

func main() {
	m.ExpectRequest(test.RequestEqualTo(fooproto.BarRequest{
		IdempotencyKey: idempotencyKey,
		SubjectId:      subjectID,
		Trigger:        fooproto.SomeFooConst,
	})).RespondWith(fooproto.BarResponse{})
}`,
			lineNumber: 8,
		},
		{
			name: "Declaration",
			src: `package main

func main() {
	fooproto.BarRequest{
		IdempotencyKey:    idempotencyKey,
		SubjectId:         subjectID,
		Trigger:           fooproto.SomeFooConst,
	}
}`,
			expected: `package main

func main() {
	m.ExpectRequest(test.RequestEqualTo(fooproto.BarRequest{
		IdempotencyKey: idempotencyKey,
		SubjectId:      subjectID,
		Trigger:        fooproto.SomeFooConst,
	})).RespondWith(fooproto.BarResponse{})
}`,
			lineNumber: 4,
		},
		{
			name: "Declaration - with single space after previous line",
			src: `package main

func main() {
	fmt.Println("hello world")

	rsp, err := fooproto.BarRequest{
		IdempotencyKey:    idempotencyKey,
		SubjectId:         subjectID,
		Trigger:           fooproto.SomeFooConst,
	}.Send(ctx).DecodeResponse()
}`,
			expected: `package main

func main() {
	fmt.Println("hello world")

	m.ExpectRequest(test.RequestEqualTo(fooproto.BarRequest{
		IdempotencyKey: idempotencyKey,
		SubjectId:      subjectID,
		Trigger:        fooproto.SomeFooConst,
	})).RespondWith(fooproto.BarResponse{})
}`,
			lineNumber: 7,
		},
		{
			name: "Declaration - with comment",
			src: `package main

func main() {
	// Here is a comment
	fooproto.BarRequest{
		IdempotencyKey:    idempotencyKey,
		SubjectId:         subjectID,
		Trigger:           fooproto.SomeFooConst,
	}
}`,
			expected: `package main

func main() {
	// Here is a comment
	m.ExpectRequest(test.RequestEqualTo(fooproto.BarRequest{
		IdempotencyKey: idempotencyKey,
		SubjectId:      subjectID,
		Trigger:        fooproto.SomeFooConst,
	})).RespondWith(fooproto.BarResponse{})
}`,
			lineNumber: 5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output, err := parseBytes([]byte(tt.src), tt.lineNumber)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, strings.Trim(output.String(), "\n"))
		})
	}
}
