package openai

// Tingly extension: admit the full content-part union in tool message content.
//
// Agent frameworks return tool screenshots as image_url parts inside
// role:"tool" messages, and OpenAI-compatible upstreams accept them. The
// upstream text-only ChatCompletionToolMessageParamContentUnion corrupted
// those parts at parse time (the image_url payload was dropped on
// re-marshal, leaving {"type":"image_url","text":""}), which upstreams
// reject with 400 invalid_parameter_error.
//
// chatcompletion.go widens OfArrayOfContentParts to
// []ChatCompletionContentPartUnionParam (the same union user message content
// uses); this file carries the source-compatibility shim for callers still
// passing []ChatCompletionContentPartTextParam to ToolMessage, plus the
// documentation of the change so it doesn't get lost on the next codegen
// pass.
//
// Fixes the ingress half of tingly-dev/tingly-box#1606.

// liftToolMessageTextParts lifts plain text parts into the content-part
// union so ToolMessage keeps accepting []ChatCompletionContentPartTextParam
// for source compatibility with the upstream SDK signature.
func liftToolMessageTextParts(parts []ChatCompletionContentPartTextParam) []ChatCompletionContentPartUnionParam {
	lifted := make([]ChatCompletionContentPartUnionParam, 0, len(parts))
	for i := range parts {
		lifted = append(lifted, ChatCompletionContentPartUnionParam{OfText: &parts[i]})
	}
	return lifted
}
