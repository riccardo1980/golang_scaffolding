package tools

import (
	"context"
	"encoding/json"
	"fmt"
)

// WeatherTool returns simulated current weather data for a requested city.
type WeatherTool struct{}

// Description returns a short human-readable description of the weather tool.
func (WeatherTool) Description() string {
	return "Returns the current weather for a given city."
}

// Run parses a JSON object containing a "city" field and returns the weather as JSON.
func (WeatherTool) Run(ctx context.Context, input string) (string, error) {
	var args struct {
		City string `json:"city"`
	}
	if err := json.Unmarshal([]byte(input), &args); err != nil {
		return "", fmt.Errorf("invalid JSON input: %w", err)
	}

	if err := sleepContext(ctx, simulatedLatency); err != nil {
		return "", err
	}

	out, err := json.Marshal(struct {
		City      string `json:"city"`
		TempC     int    `json:"temp_c"`
		Condition string `json:"condition"`
	}{
		City:      args.City,
		TempC:     18,
		Condition: "Partly Cloudy",
	})
	if err != nil {
		return "", fmt.Errorf("marshal weather response: %w", err)
	}

	return string(out), nil
}
