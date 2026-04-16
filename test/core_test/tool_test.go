package core_test

import (
	"strings"
	"testing"

	"github.com/Notailab/go-agent/agent/core"
)

func TestParseToolParamsValidatesTypesAndDefaults(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		params  string
		def     core.Parameters
		want    map[string]any
		wantErr string
	}{
		{
			name:   "applies defaults and converts types",
			params: `{"count":3}`,
			def: core.Parameters{
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
			},
			want: map[string]any{
				"name":    "anon",
				"count":   3,
				"enabled": true,
			},
		},
		{
			name:   "converts integer-like float",
			params: `{"count":3.0}`,
			def: core.Parameters{
				Type: "object",
				Properties: map[string]core.Param{
					"count": {Type: "integer"},
				},
				Required: []string{"count"},
			},
			want: map[string]any{"count": 3},
		},
		{
			name:   "rejects non integer float",
			params: `{"count":3.5}`,
			def: core.Parameters{
				Type: "object",
				Properties: map[string]core.Param{
					"count": {Type: "integer"},
				},
				Required: []string{"count"},
			},
			wantErr: "invalid type for parameter \"count\"",
		},
		{
			name:   "accepts explicit object and array values",
			params: `{"count":1,"items":[1,"two"],"config":{"mode":"fast"}}`,
			def: core.Parameters{
				Type: "object",
				Properties: map[string]core.Param{
					"count":  {Type: "integer"},
					"items":  {Type: "array"},
					"config": {Type: "object"},
				},
				Required: []string{"count"},
			},
			want: map[string]any{
				"count":  1,
				"items":  []any{float64(1), "two"},
				"config": map[string]any{"mode": "fast"},
			},
		},
		{
			name:   "rejects missing required parameter",
			params: `{}`,
			def: core.Parameters{
				Type: "object",
				Properties: map[string]core.Param{
					"count": {Type: "integer"},
				},
				Required: []string{"count"},
			},
			wantErr: "missing required parameters: [count]",
		},
		{
			name:   "rejects invalid json",
			params: `{"count":`,
			def: core.Parameters{
				Type: "object",
				Properties: map[string]core.Param{
					"count": {Type: "integer"},
				},
				Required: []string{"count"},
			},
			wantErr: "failed to parse parameters JSON:",
		},
		{
			name:   "rejects empty params string",
			params: "",
			def: core.Parameters{
				Type: "object",
				Properties: map[string]core.Param{
					"count": {Type: "integer"},
				},
				Required: []string{"count"},
			},
			wantErr: "parameters JSON is empty",
		},
		{
			name:   "rejects invalid explicit type",
			params: `{"count":"3"}`,
			def: core.Parameters{
				Type: "object",
				Properties: map[string]core.Param{
					"count": {Type: "integer"},
				},
				Required: []string{"count"},
			},
			wantErr: "invalid type for parameter \"count\"",
		},
		{
			name:   "fails when default cannot convert",
			params: `{}`,
			def: core.Parameters{
				Type: "object",
				Properties: map[string]core.Param{
					"count": {
						Type:    "integer",
						Default: "bad-default",
					},
				},
			},
			wantErr: "failed to convert parameter \"count\"",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params, err := core.ParseToolParams(tt.params, tt.def)
			if tt.wantErr != "" {
				if err == nil {
					t.Fatalf("expected error containing %q, got nil", tt.wantErr)
				}
				if !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("unexpected error: %v", err)
				}
				return
			}

			if err != nil {
				t.Fatalf("parse failed: %v", err)
			}
			if len(params) != len(tt.want) {
				t.Fatalf("unexpected param count: got %d want %d", len(params), len(tt.want))
			}
			for key, want := range tt.want {
				got, ok := params[key]
				if !ok {
					t.Fatalf("missing param %q", key)
				}
				if !deepEqualAny(got, want) {
					t.Fatalf("unexpected param %q: got %#v want %#v", key, got, want)
				}
			}
		})
	}
}

func TestParseToolParamsUsesZeroValueForNilDefaults(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		def     core.Parameters
		want    map[string]any
		wantErr string
	}{
		{
			name: "zero values for missing optional params",
			def: core.Parameters{
				Type: "object",
				Properties: map[string]core.Param{
					"name":    {Type: "string"},
					"count":   {Type: "integer"},
					"enabled": {Type: "boolean"},
					"items":   {Type: "array"},
					"config":  {Type: "object"},
				},
			},
			want: map[string]any{
				"name":    "",
				"count":   0,
				"enabled": false,
				"items":   []any{},
				"config":  map[string]any{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params, err := core.ParseToolParams(`{}`, tt.def)
			if err != nil {
				t.Fatalf("parse failed: %v", err)
			}
			for key, want := range tt.want {
				got, ok := params[key]
				if !ok {
					t.Fatalf("missing param %q", key)
				}
				if !deepEqualAny(got, want) {
					t.Fatalf("unexpected zero value for %q: got %#v want %#v", key, got, want)
				}
			}
		})
	}
}

func deepEqualAny(got, want any) bool {
	switch wantValue := want.(type) {
	case []any:
		gotValue, ok := got.([]any)
		if !ok || len(gotValue) != len(wantValue) {
			return false
		}
		for i := range wantValue {
			if !deepEqualAny(gotValue[i], wantValue[i]) {
				return false
			}
		}
		return true
	case map[string]any:
		gotValue, ok := got.(map[string]any)
		if !ok || len(gotValue) != len(wantValue) {
			return false
		}
		for key, wantItem := range wantValue {
			gotItem, ok := gotValue[key]
			if !ok || !deepEqualAny(gotItem, wantItem) {
				return false
			}
		}
		return true
	default:
		return got == want
	}
}
