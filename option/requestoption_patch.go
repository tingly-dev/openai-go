package option

import (
	"github.com/openai/openai-go/v3/internal/requestconfig"
)

// Tingly extension: lives in a sidecar file so the generated requestoption.go
// stays byte-identical to upstream.

// WithoutEnvironmentDefaults returns a RequestOption that, when passed to
// openai.NewClient, skips every ambient OPENAI_* default (OPENAI_API_KEY,
// OPENAI_ORG_ID, OPENAI_PROJECT_ID, OPENAI_BASE_URL, OPENAI_CUSTOM_HEADERS, ...).
// The regular transport defaults and the production base URL still apply.
//
// Use it for clients that talk to a provider other than the one the process
// environment is configured for, so environment credentials and headers are
// never sent to it.
//
// It exposes the internal marker the azure and bedrock clients already use.
func WithoutEnvironmentDefaults() RequestOption {
	return requestconfig.WithEnvironmentDefaultsDisabled()
}
