package ai

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strings"
)

var jsonBlockRegex = regexp.MustCompile("(?s)```(?:json)?\\s*(.+?)```")

func GenerateStruct(ctx context.Context, p AIProvider, req ChatRequest, target interface{}) error {

	schema, err := generateSchema(target)
	if err != nil {
		return fmt.Errorf("failed to generate schema: %w", err)
	}

	systemPrompt := fmt.Sprintf(`You are not a chatbot. You are a JSON data generation engine.
Your response must strictly adhere to this schema:
%s

Do not use the Markdown block (xxxjson).
Do not add comment lines.
Do not write an introductory or concluding sentence (Preamble/Postscript).
Return only raw JSON data.`, schema)

	strictSystemMsg := ChatMessage{
		Role: "system",
		Content: []Content{
			{Type: "text", Text: systemPrompt},
		},
	}

	req.Messages = append([]ChatMessage{strictSystemMsg}, req.Messages...)

	req.JSONMode = true

	resp, err := p.Generate(ctx, req)
	if err != nil {
		return err
	}

	cleanedJSON, err := sanitize(resp.Content)
	if err != nil {
		return fmt.Errorf("failed to sanitize output: %w. Raw output: %s", err, resp.Content)
	}

	if len(cleanedJSON) == 0 || cleanedJSON == "{}" {
		return fmt.Errorf("model produced empty or void JSON. Raw: %s", resp.Content)
	}

	if err := json.Unmarshal([]byte(cleanedJSON), target); err != nil {
		return fmt.Errorf("model malformed JSON produced: %w. Cleaned output: %s", err, cleanedJSON)
	}

	return nil
}

func generateSchema(target interface{}) (string, error) {
	if target == nil {
		return "", errors.New("invalid target: target cannot be nil")
	}

	t := reflect.TypeOf(target)

	if t.Kind() != reflect.Ptr {
		return "", errors.New("invalid target: pointer required (pass &struct)")
	}

	if t.Elem().Kind() != reflect.Struct {
		return "", errors.New("invalid target: pointer to struct required")
	}

	schema := make(map[string]interface{})
	schema["type"] = "object"
	schema["properties"] = generateProperties(t.Elem())

	schemaBytes, err := json.MarshalIndent(schema, "", "  ")
	if err != nil {
		return "", err
	}

	return string(schemaBytes), nil
}

func generateProperties(t reflect.Type) map[string]interface{} {
	props := make(map[string]interface{})
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		jsonTag := field.Tag.Get("json")
		if jsonTag == "-" {
			continue
		}

		name := field.Name
		if jsonTag != "" {
			parts := strings.Split(jsonTag, ",")
			name = parts[0]
		}

		desc := field.Tag.Get("description")

		prop := make(map[string]interface{})
		prop["description"] = desc

		switch field.Type.Kind() {
		case reflect.String:
			prop["type"] = "string"
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			prop["type"] = "integer"
		case reflect.Float32, reflect.Float64:
			prop["type"] = "number"
		case reflect.Bool:
			prop["type"] = "boolean"
		case reflect.Slice:
			prop["type"] = "array"
			prop["items"] = map[string]string{"type": "string"}
			if field.Type.Elem().Kind() == reflect.Struct {
				prop["items"] = map[string]interface{}{
					"type":       "object",
					"properties": generateProperties(field.Type.Elem()),
				}
			}
		case reflect.Struct:
			prop["type"] = "object"
			prop["properties"] = generateProperties(field.Type)
		default:
			prop["type"] = "string"
		}

		props[name] = prop
	}
	return props
}

func sanitize(raw string) (string, error) {
	if matches := jsonBlockRegex.FindStringSubmatch(raw); len(matches) > 1 {
		return strings.TrimSpace(matches[1]), nil
	}

	start := strings.Index(raw, "{")
	end := strings.LastIndex(raw, "}")

	if start == -1 || end == -1 || start > end {
		return "", errors.New("no JSON object found in response")
	}

	jsonStr := raw[start : end+1]

	return jsonStr, nil
}
