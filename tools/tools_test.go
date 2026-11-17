package tools

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
)

func TestWeatherTool_Run(t *testing.T) {
	tool := WeatherTool{}
	if tool.Description() == "" {
		t.Fatal("expected non-empty description")
	}

	t.Run("valid input", func(t *testing.T) {
		res, err := tool.Run(context.Background(), `{"city":"Rome"}`)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		var got struct {
			City      string `json:"city"`
			TempC     int    `json:"temp_c"`
			Condition string `json:"condition"`
		}
		if err := json.Unmarshal([]byte(res), &got); err != nil {
			t.Fatalf("output is not valid JSON %q: %v", res, err)
		}
		if got.City != "Rome" || got.TempC != 18 || got.Condition != "Partly Cloudy" {
			t.Errorf("unexpected weather payload: %+v", got)
		}
	})

	t.Run("invalid JSON", func(t *testing.T) {
		if _, err := tool.Run(context.Background(), `{bad`); err == nil {
			t.Fatal("expected error for invalid JSON, got nil")
		}
	})

	t.Run("context cancelled", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		_, err := tool.Run(ctx, `{"city":"Rome"}`)
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("got err = %v, want context.Canceled", err)
		}
	})
}

func TestCalculatorTool_Run(t *testing.T) {
	tool := CalculatorTool{}
	if tool.Description() == "" {
		t.Fatal("expected non-empty description")
	}

	t.Run("valid input", func(t *testing.T) {
		res, err := tool.Run(context.Background(), `{"a":42.5,"b":17.5}`)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if res != `{"result": 60.00}` {
			t.Errorf("got %q, want %q", res, `{"result": 60.00}`)
		}
	})

	t.Run("invalid JSON", func(t *testing.T) {
		if _, err := tool.Run(context.Background(), `{bad`); err == nil {
			t.Fatal("expected error for invalid JSON, got nil")
		}
	})

	t.Run("context cancelled", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		_, err := tool.Run(ctx, `{"a":1,"b":2}`)
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("got err = %v, want context.Canceled", err)
		}
	})
}
