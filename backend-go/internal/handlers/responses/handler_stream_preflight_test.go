package responses

import "testing"

func TestHasResponsesNonTextContent(t *testing.T) {
	t.Run("function call arguments delta", func(t *testing.T) {
		event := "event: response.function_call_arguments.delta\ndata: {\"type\":\"response.function_call_arguments.delta\",\"delta\":\"\"}\n\n"
		if !hasResponsesNonTextContent(event) {
			t.Fatal("expected function_call_arguments.delta to be treated as non-text content")
		}
	})

	t.Run("output item added function call", func(t *testing.T) {
		event := "event: response.output_item.added\ndata: {\"type\":\"response.output_item.added\",\"item\":{\"type\":\"function_call\",\"name\":\"Read\",\"call_id\":\"call_1\"}}\n\n"
		if !hasResponsesNonTextContent(event) {
			t.Fatal("expected function_call output_item to be treated as non-text content")
		}
	})

	t.Run("completed event with function call output", func(t *testing.T) {
		event := "event: response.completed\ndata: {\"type\":\"response.completed\",\"response\":{\"output\":[{\"type\":\"function_call\",\"name\":\"Read\",\"call_id\":\"call_1\",\"arguments\":\"{}\"}]}}\n\n"
		if !hasResponsesNonTextContent(event) {
			t.Fatal("expected completed event with function_call output to be treated as non-text content")
		}
	})

	t.Run("plain empty completed", func(t *testing.T) {
		event := "event: response.completed\ndata: {\"type\":\"response.completed\",\"response\":{\"output\":[]}}\n\n"
		if hasResponsesNonTextContent(event) {
			t.Fatal("did not expect empty completed event to be treated as non-text content")
		}
	})
}

func TestIsResponsesEmptyContent(t *testing.T) {
	cases := []struct {
		name  string
		text  string
		empty bool
	}{
		{name: "empty string", text: "", empty: true},
		{name: "opening brace only", text: "{", empty: true},
		{name: "whitespace brace", text: "  {  ", empty: true},
		{name: "json body", text: "{\"path\":\"/tmp/x\"}", empty: false},
		{name: "plain text", text: "hello", empty: false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := isResponsesEmptyContent(tc.text); got != tc.empty {
				t.Fatalf("isResponsesEmptyContent(%q) = %v, want %v", tc.text, got, tc.empty)
			}
		})
	}
}
