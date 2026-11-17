package tools

import (
	"context"
	"encoding/json"
	"fmt"
)

// CalculatorTool adds two floating-point numbers provided as JSON.
type CalculatorTool struct{}

// Description returns a short human-readable description of the calculator tool.
func (CalculatorTool) Description() string {
	return "Adds two numbers together."
}

// Run parses a JSON object with "a" and "b" fields and returns their sum as JSON.
func (CalculatorTool) Run(ctx context.Context, input string) (string, error) {
	var args struct {
		A float64 `json:"a"`
		B float64 `json:"b"`
	}
	if err := json.Unmarshal([]byte(input), &args); err != nil {
		return "", fmt.Errorf("invalid JSON input: %w", err)
	}

	if err := sleepContext(ctx, simulatedLatency); err != nil {
		return "", err
	}

	return fmt.Sprintf(`{"result": %.2f}`, args.A+args.B), nil
}
