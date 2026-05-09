package config

import "testing"

func TestUpstreamConfigClone_DeepCopy(t *testing.T) {
	enabled := true
	src := UpstreamConfig{
		APIKeys:                      []string{"k1"},
		ModelMapping:                 map[string]string{"a": "b"},
		CustomHeaders:                map[string]string{"X": "Y"},
		BaseURLs:                     []string{"u1"},
		FailoverRules:                []FailoverRule{{Description: "r1"}},
		NeverBlacklistKeys:           &enabled,
		RouteMessagesToResponsesPool: &enabled,
		NormalizeMetadataUserID:      &enabled,
		StripResponsesUser:           &enabled,
	}

	dst := src.Clone()
	dst.APIKeys[0] = "k2"
	dst.ModelMapping["a"] = "c"
	dst.CustomHeaders["X"] = "Z"
	dst.BaseURLs[0] = "u2"
	dst.FailoverRules[0].Description = "r2"
	*dst.NeverBlacklistKeys = false
	*dst.RouteMessagesToResponsesPool = false
	*dst.NormalizeMetadataUserID = false
	*dst.StripResponsesUser = false

	if src.APIKeys[0] != "k1" {
		t.Fatalf("APIKeys shared backing array")
	}
	if src.ModelMapping["a"] != "b" {
		t.Fatalf("ModelMapping shared backing map")
	}
	if src.CustomHeaders["X"] != "Y" {
		t.Fatalf("CustomHeaders shared backing map")
	}
	if src.BaseURLs[0] != "u1" {
		t.Fatalf("BaseURLs shared backing array")
	}
	if src.FailoverRules[0].Description != "r1" {
		t.Fatalf("FailoverRules shared backing array")
	}
	if !*src.NeverBlacklistKeys || !*src.RouteMessagesToResponsesPool || !*src.NormalizeMetadataUserID || !*src.StripResponsesUser {
		t.Fatalf("boolean pointer fields were not deep copied")
	}
}

func TestSupportsModel(t *testing.T) {
	tests := []struct {
		name            string
		supportedModels []string
		model           string
		want            bool
	}{
		{"空列表匹配所有", nil, "gpt-4o", true},
		{"空列表匹配空模型", nil, "", true},
		{"精确匹配", []string{"gpt-4o"}, "gpt-4o", true},
		{"精确不匹配", []string{"gpt-4o"}, "gpt-4-turbo", false},
		{"通配符匹配", []string{"gpt-4*"}, "gpt-4o", true},
		{"通配符匹配turbo", []string{"gpt-4*"}, "gpt-4-turbo", true},
		{"通配符不匹配", []string{"gpt-4*"}, "o3", false},
		{"多模式匹配第一个", []string{"gpt-4*", "claude-*"}, "gpt-4o", true},
		{"多模式匹配第二个", []string{"gpt-4*", "claude-*"}, "claude-3-opus", true},
		{"多模式都不匹配", []string{"gpt-4*", "claude-*"}, "o3", false},
		{"精确和通配符混合", []string{"o3", "gpt-4*"}, "o3", true},
		{"通配符星号本身", []string{"*"}, "anything", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := &UpstreamConfig{SupportedModels: tt.supportedModels}
			if got := u.SupportsModel(tt.model); got != tt.want {
				t.Errorf("SupportsModel(%q) = %v, want %v", tt.model, got, tt.want)
			}
		})
	}
}

func TestResolveReasoningEffort(t *testing.T) {
	upstream := &UpstreamConfig{
		ReasoningMapping: map[string]string{
			"gpt-5":         "high",
			"gpt-5.1-codex": "xhigh",
			"o3":            "medium",
		},
	}

	tests := []struct {
		name  string
		model string
		want  string
	}{
		{"精确匹配", "o3", "medium"},
		{"最长匹配优先", "gpt-5.1-codex", "xhigh"},
		{"模糊匹配回退", "gpt-5.1", "xhigh"},
		{"未匹配返回空", "claude-3-7-sonnet", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ResolveReasoningEffort(tt.model, upstream); got != tt.want {
				t.Fatalf("ResolveReasoningEffort(%q) = %q, want %q", tt.model, got, tt.want)
			}
		})
	}
}
