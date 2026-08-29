package openai

import (
	"encoding/json"

	"github.com/openai/openai-go/v3/packages/respjson"
)

func chatCompletionMessageToAssistantParam(r ChatCompletionMessage) ChatCompletionAssistantMessageParam {
	var p ChatCompletionAssistantMessageParam

	// It is important to not rely on the JSON metadata property
	// here, it may be unset if the receiver was generated via a
	// [ChatCompletionAccumulator].
	//
	// Explicit null is intentionally elided from the response.
	if r.Content != "" {
		p.Content.OfString = String(r.Content)
	}
	if r.Refusal != "" {
		p.Refusal = String(r.Refusal)
	}

	p.Audio.ID = r.Audio.ID
	p.Role = r.Role
	p.FunctionCall.Arguments = r.FunctionCall.Arguments
	p.FunctionCall.Name = r.FunctionCall.Name

	if len(r.ToolCalls) > 0 {
		for _, v := range r.ToolCalls {
			u := ChatCompletionMessageToolCallUnionParam{}
			switch v.AsAny().(type) {
			case ChatCompletionMessageFunctionToolCall:
				u.OfFunction = &ChatCompletionMessageFunctionToolCallParam{
					ID: v.ID,
					Function: ChatCompletionMessageFunctionToolCallFunctionParam{
						Arguments: v.Function.Arguments,
						Name:      v.Function.Name,
					},
				}
			case ChatCompletionMessageCustomToolCall:
				u.OfCustom = &ChatCompletionMessageCustomToolCallParam{
					ID: v.ID,
					Custom: ChatCompletionMessageCustomToolCallCustomParam{
						Input: v.Custom.Input,
						Name:  v.Custom.Name,
					},
				}
			}

			p.ToolCalls = append(p.ToolCalls, u)
		}
	}
	return p
}

// extraFieldsToAny converts ExtraFields from response type to a format suitable for request params.
//
// Note: unknown/extra fields in respjson.Field are ALWAYS marked status=invalid
// ("couldn't be marshaled to a known type", raw JSON preserved) — Field.Valid()
// is false for every real extra field, so we must NOT filter on Valid().
// Filter on raw presence instead, skipping explicit nulls.
func extraFieldsToAny(extraFields map[string]respjson.Field) map[string]any {
	if len(extraFields) == 0 {
		return nil
	}
	result := make(map[string]any, len(extraFields))
	for k, v := range extraFields {
		raw := v.Raw()
		if raw == "" || raw == respjson.Null {
			continue
		}
		var val any
		if err := json.Unmarshal([]byte(raw), &val); err == nil {
			result[k] = val
		}
	}
	return result
}

func (r ChatCompletionMessage) ToParam() ChatCompletionMessageParamUnion {
	asst := r.ToAssistantMessageParam()
	return ChatCompletionMessageParamUnion{OfAssistant: &asst}
}

func (r ChatCompletionMessage) ToAssistantMessageParam() ChatCompletionAssistantMessageParam {
	p := chatCompletionMessageToAssistantParam(r)

	// Copy ExtraFields from the response to the param type
	if extras := extraFieldsToAny(r.JSON.ExtraFields); len(extras) > 0 {
		p.SetExtraFields(extras)
	}

	return p
}

func AssistantMessage[T string | []ChatCompletionAssistantMessageParamContentArrayOfContentPartUnion](content T) ChatCompletionMessageParamUnion {
	return ChatCompletionMessageParamOfAssistant(content)
}
