package core_test

import (
	"testing"

	"github.com/Notailab/go-agent/agent/core"
)

func TestParseToolParamsValidatesTypesAndDefaults(t *testing.T) {
	t.Parallel()

	paramsDef := core.Parameters{
		Type: "object",
		Properties: map[string]core.Param{
			"name": {
				Type:    "string",
				Default: "anon",
			},
			"count": {
				Type: "integer",
			},
			"enabled": {
				Type:    "boolean",
				Default: true,
			},
		},
		Required: []string{"count"},
	}

	params, err := core.ParseToolParams(`{"count":3}`, paramsDef)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	if got := params["name"]; got != "anon" {
		t.Fatalf("unexpected default name: %#v", got)
	}
	if got := params["count"]; got != 3 {
		t.Fatalf("unexpected count: %#v", got)
	}
	if got := params["enabled"]; got != true {
		t.Fatalf("unexpected default enabled: %#v", got)
	}

	if _, err := core.ParseToolParams(`{"count":"3"}`, paramsDef); err == nil {
		t.Fatal("expected type validation error")
	}
}

func TestParseToolParamsUsesZeroValueForNilDefaults(t *testing.T) {
	t.Parallel()

	paramsDef := core.Parameters{
		Type: "object",
		Properties: map[string]core.Param{
			"name": {
				Type: "string",
			},
			"count": {
				Type: "integer",
			},
			"enabled": {
				Type: "boolean",
			},
			"items": {
				Type: "array",
			},
			"config": {
				Type: "object",
			},
		},
	}

	params, err := core.ParseToolParams(`{}`, paramsDef)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	if got := params["name"]; got != "" {
		t.Fatalf("unexpected string zero value: %#v", got)
	}
	if got := params["count"]; got != 0 {
		t.Fatalf("unexpected integer zero value: %#v", got)
	}
	if got := params["enabled"]; got != false {
		t.Fatalf("unexpected boolean zero value: %#v", got)
	}
	if got := params["items"]; len(got.([]any)) != 0 {
		t.Fatalf("unexpected array zero value: %#v", got)
	}
	if got := params["config"]; len(got.(map[string]any)) != 0 {
		t.Fatalf("unexpected object zero value: %#v", got)
	}
}
