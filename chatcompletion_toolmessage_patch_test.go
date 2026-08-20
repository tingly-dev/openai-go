package openai_test

import (
	"encoding/json"
	"testing"

	"github.com/openai/openai-go/v3"
)

// TestToolMessageContentUnionImagePartRoundTrip locks the tingly extension on
// ChatCompletionToolMessageParamContentUnion: role:"tool" message content
// arrays admit the full content-part union, so image_url parts (tool
// screenshots returned by agent frameworks) survive unmarshal → marshal
// instead of degrading to {"type":"image_url","text":""}.
func TestToolMessageContentUnionImagePartRoundTrip(t *testing.T) {
	const imageURL = "data:image/png;base64,iVBORw0KGgo="
	raw := []byte(`{
		"role": "tool",
		"tool_call_id": "call_1",
		"content": [
			{"type": "text", "text": "Image loaded."},
			{"type": "image_url", "image_url": {"url": "` + imageURL + `"}}
		]
	}`)

	var msg openai.ChatCompletionMessageParamUnion
	if err := json.Unmarshal(raw, &msg); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if msg.OfTool == nil {
		t.Fatal("expected tool variant")
	}
	parts := msg.OfTool.Content.OfArrayOfContentParts
	if len(parts) != 2 {
		t.Fatalf("expected 2 content parts, got %d", len(parts))
	}
	if parts[0].OfText == nil || parts[0].OfText.Text != "Image loaded." {
		t.Fatalf("text part mismatch: %+v", parts[0])
	}
	if parts[1].OfImageURL == nil || parts[1].OfImageURL.ImageURL.URL != imageURL {
		t.Fatalf("image part mismatch: %+v", parts[1])
	}

	out, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var m map[string]any
	if err := json.Unmarshal(out, &m); err != nil {
		t.Fatalf("re-unmarshal: %v", err)
	}
	content, ok := m["content"].([]any)
	if !ok || len(content) != 2 {
		t.Fatalf("content should round-trip as a 2-part array, got: %v", m["content"])
	}
	img, _ := content[1].(map[string]any)
	if img["type"] != "image_url" {
		t.Fatalf("second part should stay image_url, got: %v", img)
	}
	imgURL, ok := img["image_url"].(map[string]any)
	if !ok || imgURL["url"] != imageURL {
		t.Fatalf("image_url.url payload must survive, got: %v", img)
	}
	if _, hasText := img["text"]; hasText {
		t.Fatalf("image part must not grow a text field, got: %v", img)
	}
}

// TestToolMessageConstructorVariants locks the ToolMessage constructor's three
// accepted content forms.
func TestToolMessageConstructorVariants(t *testing.T) {
	if msg := openai.ToolMessage("plain", "call_1"); msg.OfTool.Content.OfString.Value != "plain" {
		t.Fatal("string form should populate OfString")
	}

	textParts := []openai.ChatCompletionContentPartTextParam{{Text: "a"}, {Text: "b"}}
	msg := openai.ToolMessage(textParts, "call_1")
	parts := msg.OfTool.Content.OfArrayOfContentParts
	if len(parts) != 2 || parts[0].OfText == nil || parts[0].OfText.Text != "a" {
		t.Fatalf("text-part form should lift into the union, got: %+v", parts)
	}

	unionParts := []openai.ChatCompletionContentPartUnionParam{
		{OfImageURL: &openai.ChatCompletionContentPartImageParam{
			ImageURL: openai.ChatCompletionContentPartImageImageURLParam{URL: "https://example.com/x.png"},
		}},
	}
	msg = openai.ToolMessage(unionParts, "call_1")
	parts = msg.OfTool.Content.OfArrayOfContentParts
	if len(parts) != 1 || parts[0].OfImageURL == nil {
		t.Fatalf("union form should pass through, got: %+v", parts)
	}
}
