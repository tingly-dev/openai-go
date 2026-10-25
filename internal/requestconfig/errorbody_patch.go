package requestconfig

import (
	"encoding/json"
	"strings"

	"github.com/tidwall/gjson"
)

// unwrapErrorBody returns the JSON object to decode into apierror.Error.
//
// OpenAI wraps errors as {"error": {...}}, but compatible upstreams (Codex,
// proxies, gateways) also send {"error": "text"}, {"detail": "text"},
// {"message": "text"} or a non-JSON body. Those must not turn into a decode
// failure that hides the real HTTP status, so the text is folded into a
// message-only error object.
func unwrapErrorBody(contents []byte) []byte {
	if inner := gjson.GetBytes(contents, "error"); inner.IsObject() {
		return []byte(inner.Raw)
	}
	text := ""
	for _, path := range []string{"error", "detail", "message"} {
		if v := gjson.GetBytes(contents, path); v.Type == gjson.String && v.String() != "" {
			text = v.String()
			break
		}
	}
	if text == "" {
		if v := gjson.GetBytes(contents, "detail"); v.Exists() {
			text = v.Raw
		}
	}
	if text == "" {
		text = strings.TrimSpace(string(contents))
	}
	out, err := json.Marshal(map[string]string{"message": text})
	if err != nil {
		return []byte(`{}`)
	}
	return out
}
