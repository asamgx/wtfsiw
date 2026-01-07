package tools

import (
	"github.com/sashabaranov/go-openai"
	"github.com/sashabaranov/go-openai/jsonschema"
)

// ToOpenAITools converts tool definitions to OpenAI tool format
func ToOpenAITools(tools []ToolDefinition) []openai.Tool {
	result := make([]openai.Tool, len(tools))
	for i, tool := range tools {
		result[i] = openai.Tool{
			Type: openai.ToolTypeFunction,
			Function: &openai.FunctionDefinition{
				Name:        tool.Name,
				Description: tool.Description,
				Parameters:  toOpenAISchema(tool.Parameters),
			},
		}
	}
	return result
}

func toOpenAISchema(params []ToolParameter) jsonschema.Definition {
	properties := make(map[string]jsonschema.Definition)
	required := make([]string, 0)

	for _, p := range params {
		properties[p.Name] = paramToJSONSchema(p)
		if p.Required {
			required = append(required, p.Name)
		}
	}

	return jsonschema.Definition{
		Type:       jsonschema.Object,
		Properties: properties,
		Required:   required,
	}
}

func paramToJSONSchema(p ToolParameter) jsonschema.Definition {
	def := jsonschema.Definition{
		Description: p.Description,
	}

	switch p.Type {
	case "string":
		def.Type = jsonschema.String
		if len(p.Enum) > 0 {
			def.Enum = p.Enum
		}
	case "integer":
		def.Type = jsonschema.Integer
	case "number":
		def.Type = jsonschema.Number
	case "boolean":
		def.Type = jsonschema.Boolean
	case "array":
		def.Type = jsonschema.Array
		if p.Items != nil {
			itemDef := paramToJSONSchema(*p.Items)
			def.Items = &itemDef
		}
	case "object":
		def.Type = jsonschema.Object
	}

	return def
}

// ToAnthropicInputSchema converts tool parameters to Anthropic input schema format
// Returns a map that can be used as InputSchema in anthropic.ToolParam
func ToAnthropicInputSchema(params []ToolParameter) map[string]interface{} {
	properties := make(map[string]interface{})
	required := make([]string, 0)

	for _, p := range params {
		properties[p.Name] = paramToAnthropicSchema(p)
		if p.Required {
			required = append(required, p.Name)
		}
	}

	schema := map[string]interface{}{
		"type":       "object",
		"properties": properties,
	}

	if len(required) > 0 {
		schema["required"] = required
	}

	return schema
}

func paramToAnthropicSchema(p ToolParameter) map[string]interface{} {
	schema := map[string]interface{}{
		"type":        p.Type,
		"description": p.Description,
	}

	if len(p.Enum) > 0 {
		schema["enum"] = p.Enum
	}

	if p.Type == "array" && p.Items != nil {
		schema["items"] = paramToAnthropicSchema(*p.Items)
	}

	return schema
}
