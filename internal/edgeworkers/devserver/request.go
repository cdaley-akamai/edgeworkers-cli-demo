package devserver

import (
	"crypto/rand"
	_ "embed"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/santhosh-tekuri/jsonschema/v6"
)

//go:embed invocation.schema.json
var invocationSchemaFile []byte

var (
	invocationSchemaOnce sync.Once
	invocationSchema     *jsonschema.Schema
	invocationSchemaErr  error
)

const requestIDChars = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"

// generateRequestID returns an 8-character alphanumeric identifier for invocations that
// omit requestId.
func generateRequestID() (string, error) {
	raw := make([]byte, 8)
	if _, err := rand.Read(raw); err != nil {
		return "", fmt.Errorf("generate request ID: %w", err)
	}
	id := make([]byte, 8)
	for i, b := range raw {
		id[i] = requestIDChars[int(b)%len(requestIDChars)]
	}
	return string(id), nil
}

// handleRequest validates body against the invocation schema (edgeWorkerId, eventHandler,
// settings, request, and response) and returns the payload as compacted JSON.
//
// requestId is optional; when absent, handleRequest assigns a generated 8-character
// alphanumeric ID before validation so every invocation is uniquely correlatable in logs.
//
// The scaffold echoes the request back unchanged; invoking the EdgeWorker code with the
// validated request/response objects is future work.
func handleRequest(body []byte) (json.RawMessage, error) {
	var payload any
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, fmt.Errorf("invalid JSON request: %w", err)
	}
	if object, ok := payload.(map[string]any); ok {
		if _, present := object["requestId"]; !present {
			requestID, err := generateRequestID()
			if err != nil {
				return nil, err
			}
			object["requestId"] = requestID
		}
	}
	schema, err := compiledInvocationSchema()
	if err != nil {
		return nil, err
	}
	if err := schema.Validate(payload); err != nil {
		return nil, fmt.Errorf("validate invocation JSON: %w", err)
	}
	compact, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("compact JSON request: %w", err)
	}
	return compact, nil
}

func compiledInvocationSchema() (*jsonschema.Schema, error) {
	invocationSchemaOnce.Do(func() {
		var schemaDocument any
		if err := json.Unmarshal(invocationSchemaFile, &schemaDocument); err != nil {
			invocationSchemaErr = fmt.Errorf("parse embedded invocation schema: %w", err)
			return
		}
		compiler := jsonschema.NewCompiler()
		if err := compiler.AddResource("invocation.schema.json", schemaDocument); err != nil {
			invocationSchemaErr = fmt.Errorf("load embedded invocation schema: %w", err)
			return
		}
		invocationSchema, invocationSchemaErr = compiler.Compile("invocation.schema.json")
		if invocationSchemaErr != nil {
			invocationSchemaErr = fmt.Errorf("compile embedded invocation schema: %w", invocationSchemaErr)
		}
	})
	return invocationSchema, invocationSchemaErr
}
