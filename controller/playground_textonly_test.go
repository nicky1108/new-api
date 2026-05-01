package controller

import (
	"encoding/json"
	"testing"
)

func TestBuildTextOnlyPlaygroundChatBodyStripsMediaParts(t *testing.T) {
	body := []byte(`{
		"model": "deepseek-chat",
		"messages": [
			{"role": "user", "content": "hello"},
			{"role": "assistant", "content": [{"type": "image_url", "image_url": {"url": "https://example.test/image.png"}}]},
			{"role": "user", "content": [{"type": "text", "text": "describe this"}, {"type": "image_url", "image_url": {"url": "data:image/png;base64,abc"}}]}
		],
		"stream": true,
		"metadata": {"source": "playground"}
	}`)

	normalized, changed, err := buildTextOnlyPlaygroundChatBody(body)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !changed {
		t.Fatal("expected body to be changed")
	}

	var payload map[string]any
	if err := json.Unmarshal(normalized, &payload); err != nil {
		t.Fatalf("normalized body is invalid JSON: %v", err)
	}

	messages, ok := payload["messages"].([]any)
	if !ok {
		t.Fatalf("messages missing from normalized body: %#v", payload)
	}
	if len(messages) != 2 {
		t.Fatalf("expected media-only message to be dropped, got %d messages: %#v", len(messages), messages)
	}

	userMessage, ok := messages[1].(map[string]any)
	if !ok {
		t.Fatalf("expected second message object, got %#v", messages[1])
	}
	if userMessage["content"] != "describe this" {
		t.Fatalf("expected image content to become text, got %#v", userMessage["content"])
	}
	if _, exists := payload["metadata"]; !exists {
		t.Fatalf("expected unrelated payload fields to be preserved: %#v", payload)
	}
}

func TestBuildTextOnlyPlaygroundChatBodyLeavesTextMessagesUnchanged(t *testing.T) {
	body := []byte(`{"model":"deepseek-chat","messages":[{"role":"user","content":"hello"}],"stream":false}`)

	normalized, changed, err := buildTextOnlyPlaygroundChatBody(body)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if changed {
		t.Fatalf("expected plain text body to be unchanged, got: %s", string(normalized))
	}
}
