package tools

// ToolDefinition represents a tool that the AI can call
type ToolDefinition struct {
	Name        string
	Description string
	Parameters  []ToolParameter
}

// ToolParameter defines a single parameter for a tool
type ToolParameter struct {
	Name        string
	Type        string         // "string", "integer", "number", "boolean", "array", "object"
	Description string
	Required    bool
	Enum        []string       // optional: constrained values
	Items       *ToolParameter // for arrays: type of items
}

// ToolCall represents a request from the AI to execute a tool
type ToolCall struct {
	ID        string
	Name      string
	Arguments map[string]interface{}
}

// ToolResult represents the result of executing a tool
type ToolResult struct {
	ToolCallID string
	Content    string
	IsError    bool
}

// Helper methods for extracting typed arguments

// GetString extracts a string argument
func (tc *ToolCall) GetString(key string) string {
	if v, ok := tc.Arguments[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

// GetInt extracts an integer argument
func (tc *ToolCall) GetInt(key string) int {
	if v, ok := tc.Arguments[key]; ok {
		switch n := v.(type) {
		case float64:
			return int(n)
		case int:
			return n
		}
	}
	return 0
}

// GetFloat extracts a float argument
func (tc *ToolCall) GetFloat(key string) float64 {
	if v, ok := tc.Arguments[key]; ok {
		if f, ok := v.(float64); ok {
			return f
		}
	}
	return 0
}

// GetBool extracts a boolean argument
func (tc *ToolCall) GetBool(key string) bool {
	if v, ok := tc.Arguments[key]; ok {
		if b, ok := v.(bool); ok {
			return b
		}
	}
	return false
}

// GetStringArray extracts a string array argument
func (tc *ToolCall) GetStringArray(key string) []string {
	if v, ok := tc.Arguments[key]; ok {
		if arr, ok := v.([]interface{}); ok {
			result := make([]string, 0, len(arr))
			for _, item := range arr {
				if s, ok := item.(string); ok {
					result = append(result, s)
				}
			}
			return result
		}
	}
	return nil
}
