package openai_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
)

func TestWithoutEnvironmentDefaults(t *testing.T) {
	t.Setenv("OPENAI_API_KEY", "sk-env")
	t.Setenv("OPENAI_ORG_ID", "org-env")
	t.Setenv("OPENAI_PROJECT_ID", "proj-env")
	t.Setenv("OPENAI_CUSTOM_HEADERS", "X-Env-Leak: 1")

	run := func(extra ...option.RequestOption) http.Header {
		var got http.Header
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			got = r.Header.Clone()
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{}`))
		}))
		defer srv.Close()
		opts := append([]option.RequestOption{option.WithBaseURL(srv.URL), option.WithMaxRetries(0)}, extra...)
		c := openai.NewClient(opts...)
		_, _ = c.Chat.Completions.New(context.Background(), openai.ChatCompletionNewParams{
			Model:    "m",
			Messages: []openai.ChatCompletionMessageParamUnion{openai.UserMessage("hi")},
		})
		return got
	}

	// Baseline: without the option, ambient env leaks into the request.
	h := run()
	if h.Get("Authorization") != "Bearer sk-env" || h.Get("OpenAI-Organization") != "org-env" || h.Get("X-Env-Leak") != "1" {
		t.Fatalf("baseline: expected env headers, got %v", h)
	}

	// With the option, no env value is sent; explicit options still apply.
	h = run(option.WithoutEnvironmentDefaults(), option.WithAPIKey("k"))
	for _, k := range []string{"OpenAI-Organization", "OpenAI-Project", "X-Env-Leak"} {
		if v := h.Get(k); v != "" {
			t.Errorf("%s = %q, want empty", k, v)
		}
	}
	if got := h.Get("Authorization"); got != "Bearer k" {
		t.Errorf("Authorization = %q, want %q", got, "Bearer k")
	}
}
