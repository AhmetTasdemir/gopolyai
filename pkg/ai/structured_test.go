package ai

import (
	"testing"
)

type TestStruct struct {
	Name string `json:"name" description:"The name of the person"`
	Age  int    `json:"age" description:"The age of the person"`
}

func TestGenerateSchema(t *testing.T) {
	target := &TestStruct{}
	schema, err := generateSchema(target)
	if err != nil {
		t.Fatalf("generateSchema failed: %v", err)
	}

	t.Logf("Generated Schema: %s", schema)

	if len(schema) == 0 {
		t.Error("Schema is empty")
	}
}

func TestSanitize(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{
			name:  "Clean JSON",
			input: `{"key": "value"}`,
			want:  `{"key": "value"}`,
		},
		{
			name:  "JSON with Markdown",
			input: "```json\n{\"key\": \"value\"}\n```",
			want:  `{"key": "value"}`,
		},
		{
			name:  "JSON with Text",
			input: "Here is the JSON: {\"key\": \"value\"}",
			want:  `{"key": "value"}`,
		},
		{
			name:    "Invalid JSON",
			input:   "No JSON here",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := sanitize(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("sanitize() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("sanitize() = %v, want %v", got, tt.want)
			}
		})
	}
}
