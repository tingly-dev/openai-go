package openai

// Regression tests for tingly extras handling. These pin the behavior that
// tingly-box depends on (see tingly-box commit ed4ff1bdd: DeepSeek 400
// "reasoning_content in the thinking mode must be passed back" was caused by
// union-level ExtraFields being dropped by MarshalUnion; the product migrated
// to variant-level (OfAssistant) extras, which serialize natively).
//
// They also pin the tingly-dev patch behavior:
//   - ToParam()/ToAssistantMessageParam() copies unknown response fields
//     (tingly patch + fix: respjson marks unknown fields status=invalid, so
//     filtering on Field.Valid() dropped them all).
//
// Union-level extras are intentionally NOT serialized (upstream MarshalUnion
// behavior, unchanged by tingly-dev since the union MarshalJSON override was
// dropped at v3.54.0) — TestUnionLevelExtrasDropped documents this so nobody
// reintroduces the DeepSeek 400 class of bug by relying on it.

import (
	"encoding/json"
	"strings"
	"testing"
)

// All six variants must serialize variant-level (inner) extras natively.
// This is the path tingly-box uses (OfAssistant.SetExtraFields); it does not
// involve any tingly patch.
func TestChatCompletionExtrasRegressionVariantMatrix(t *testing.T) {
	cases := []struct {
		name string
		mk   func() ChatCompletionMessageParamUnion
	}{
		{"developer", func() ChatCompletionMessageParamUnion {
			p := ChatCompletionDeveloperMessageParam{}
			p.Content.OfString = String("d")
			p.SetExtraFields(map[string]any{"x_cust": "dev"})
			return ChatCompletionMessageParamUnion{OfDeveloper: &p}
		}},
		{"system", func() ChatCompletionMessageParamUnion {
			p := ChatCompletionSystemMessageParam{}
			p.Content.OfString = String("s")
			p.SetExtraFields(map[string]any{"x_cust": "sys"})
			return ChatCompletionMessageParamUnion{OfSystem: &p}
		}},
		{"user", func() ChatCompletionMessageParamUnion {
			p := ChatCompletionUserMessageParam{}
			p.Content.OfString = String("u")
			p.SetExtraFields(map[string]any{"x_cust": "usr"})
			return ChatCompletionMessageParamUnion{OfUser: &p}
		}},
		{"assistant", func() ChatCompletionMessageParamUnion {
			p := ChatCompletionAssistantMessageParam{}
			p.Content.OfString = String("a")
			p.SetExtraFields(map[string]any{"x_cust": "asst"})
			return ChatCompletionMessageParamUnion{OfAssistant: &p}
		}},
		{"tool", func() ChatCompletionMessageParamUnion {
			p := ChatCompletionToolMessageParam{}
			p.Content.OfString = String("t")
			p.SetExtraFields(map[string]any{"x_cust": "tool"})
			return ChatCompletionMessageParamUnion{OfTool: &p}
		}},
		{"function", func() ChatCompletionMessageParamUnion {
			p := ChatCompletionFunctionMessageParam{}
			p.Name = "f"
			p.SetExtraFields(map[string]any{"x_cust": "fn"})
			return ChatCompletionMessageParamUnion{OfFunction: &p}
		}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			b, err := json.Marshal(c.mk())
			if err != nil {
				t.Fatalf("marshal: %v", err)
			}
			if !strings.Contains(string(b), `"x_cust"`) {
				t.Fatalf("variant-level extras lost for %s: %s", c.name, b)
			}
		})
	}
}

// ToParam() must round-trip unknown response fields (e.g. DeepSeek
// reasoning_content) into the request param. Guards the tingly patch +
// the extraFieldsToAny fix (Valid()-filter used to drop every unknown
// field because respjson marks them status=invalid).
func TestChatCompletionExtrasRegressionRoundTrip(t *testing.T) {
	raw := `{"content":"hi","role":"assistant","x_thinking":{"budget":128},"reasoning_content":"because"}`
	var msg ChatCompletionMessage
	if err := json.Unmarshal([]byte(raw), &msg); err != nil {
		t.Fatal(err)
	}
	b, err := json.Marshal(msg.ToParam())
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{`"x_thinking"`, `"reasoning_content"`} {
		if !strings.Contains(string(b), want) {
			t.Fatalf("round-trip lost %s: %s", want, b)
		}
	}
}

// Explicit JSON null extras must NOT be copied back (they would serialize as
// explicit nulls in the request).
func TestChatCompletionExtrasRegressionSkipsNulls(t *testing.T) {
	raw := `{"content":"hi","role":"assistant","x_nullable":null,"x_kept":1}`
	var msg ChatCompletionMessage
	if err := json.Unmarshal([]byte(raw), &msg); err != nil {
		t.Fatal(err)
	}
	b, err := json.Marshal(msg.ToParam())
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(b), `"x_nullable"`) {
		t.Fatalf("explicit null should be skipped: %s", b)
	}
	if !strings.Contains(string(b), `"x_kept"`) {
		t.Fatalf("non-null extra should be kept: %s", b)
	}
}

// DOCUMENTATION TEST: union-level extras are NOT serialized (upstream
// MarshalUnion behavior). Do not "fix" call sites by setting extras on the
// union — set them on the variant (OfAssistant etc.) instead. This is the
// exact mistake that caused the DeepSeek 400 errors (tingly-box ed4ff1bdd).
func TestChatCompletionExtrasRegressionUnionLevelDropped(t *testing.T) {
	p := ChatCompletionAssistantMessageParam{}
	p.Content.OfString = String("a")
	u := ChatCompletionMessageParamUnion{OfAssistant: &p}
	u.SetExtraFields(map[string]any{"x_union_level": "dropped"})
	b, err := json.Marshal(u)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(b), `"x_union_level"`) {
		t.Fatalf("union-level extras unexpectedly serialized: %s — if this changed upstream, tingly-box can drop its union-level workarounds", b)
	}
}
