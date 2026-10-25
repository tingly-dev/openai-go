package requestconfig

import "testing"

func TestUnwrapErrorBody(t *testing.T) {
	cases := map[string]struct{ in, want string }{
		"openai envelope": {`{"error":{"message":"m","type":"t"}}`, `{"message":"m","type":"t"}`},
		"string error":    {`{"error":"boom"}`, `{"message":"boom"}`},
		"detail string":   {`{"detail":"Not Found"}`, `{"message":"Not Found"}`},
		"detail object":   {`{"detail":[{"msg":"x"}]}`, `{"message":"[{\"msg\":\"x\"}]"}`},
		"top message":     {`{"message":"nope"}`, `{"message":"nope"}`},
		"plain text":      {"  bad gateway \n", `{"message":"bad gateway"}`},
		"empty":           {``, `{"message":""}`},
	}
	for name, c := range cases {
		if got := string(unwrapErrorBody([]byte(c.in))); got != c.want {
			t.Errorf("%s: got %s want %s", name, got, c.want)
		}
	}
}
