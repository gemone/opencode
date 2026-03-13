// Package framework provides parameter schema definition
package framework

import (
	"encoding/json"
	"fmt"
)

// Schema defines the structure of tool parameters
type Schema struct {
	Properties   map[string]Property `json:"properties"`
	Required     []string             `json:"required,omitempty"`
	AllowUnknown bool                 `json:"allow_unknown,omitempty"`
}

// Property defines a single parameter property
type Property struct {
	Type        string   `json:"type"`
	Description string   `json:"description,omitempty"`
	Default     any      `json:"default,omitempty"`
	Required    bool     `json:"required"`
	Enum        []string `json:"enum,omitempty"`
	Min         *float64 `json:"min,omitempty"`
	Max         *float64 `json:"max,omitempty"`
	Pattern     string   `json:"pattern,omitempty"`
	Format      string   `json:"format,omitempty"`
	Items       *Property `json:"items,omitempty"` // For array types
}

// NewSchema creates a new parameter schema
func NewSchema() *Schema {
	return &Schema{
		Properties:   make(map[string]Property),
		Required:     make([]string, 0),
		AllowUnknown: false,
	}
}

// AddProperty adds a property to the schema
func (s *Schema) AddProperty(name string, prop Property) *Schema {
	s.Properties[name] = prop
	if prop.Required {
		s.Required = append(s.Required, name)
	}
	return s
}

// StringProperty creates a string property
func StringProperty(description string) Property {
	return Property{
		Type:        "string",
		Description: description,
	}
}

// NumberProperty creates a number property
func NumberProperty(description string) Property {
	return Property{
		Type:        "number",
		Description: description,
	}
}

// IntegerProperty creates an integer property
func IntegerProperty(description string) Property {
	return Property{
		Type:        "integer",
		Description: description,
	}
}

// BooleanProperty creates a boolean property
func BooleanProperty(description string) Property {
	return Property{
		Type:        "boolean",
		Description: description,
	}
}

// ArrayProperty creates an array property
func ArrayProperty(description string, items *Property) Property {
	return Property{
		Type:        "array",
		Description: description,
		Items:       items,
	}
}

// ObjectProperty creates an object property
func ObjectProperty(description string) Property {
	return Property{
		Type:        "object",
		Description: description,
	}
}

// WithDefault sets the default value for a property
func (p Property) WithDefault(defaultValue any) Property {
	p.Default = defaultValue
	return p
}

// WithRequired marks a property as required
func (p Property) WithRequired() Property {
	p.Required = true
	return p
}

// WithEnum sets the enum values for a property
func (p Property) WithEnum(values ...string) Property {
	p.Enum = values
	return p
}

// WithMin sets the minimum value for a number property
func (p Property) WithMin(min float64) Property {
	p.Min = &min
	return p
}

// WithMax sets the maximum value for a number property
func (p Property) WithMax(max float64) Property {
	p.Max = &max
	return p
}

// WithPattern sets the regex pattern for a string property
func (p Property) WithPattern(pattern string) Property {
	p.Pattern = pattern
	return p
}

// WithFormat sets the format for a property
func (p Property) WithFormat(format string) Property {
	p.Format = format
	return p
}

// MarshalJSON implements custom JSON marshaling for Schema
func (s Schema) MarshalJSON() ([]byte, error) {
	type Alias Schema
	return json.Marshal(struct {
		Type string `json:"type"`
		Alias
	}{
		Type: "object",
		Alias: (Alias)(s),
	})
}

// ValidateEnum checks if a value is valid against an enum
func ValidateEnum(value string, enum []string) error {
	for _, valid := range enum {
		if value == valid {
			return nil
		}
	}
	return fmt.Errorf("value must be one of: %v", enum)
}

// ValidateRange checks if a number is within range
func ValidateRange(value float64, min, max *float64) error {
	if min != nil && value < *min {
		return fmt.Errorf("value must be >= %v", *min)
	}
	if max != nil && value > *max {
		return fmt.Errorf("value must be <= %v", *max)
	}
	return nil
}

// Common schemas for tool parameters

// FilePathSchema returns a schema for file path parameters
func FilePathSchema() Property {
	return StringProperty("Absolute or relative path to a file").WithFormat("uri")
}

// DirectoryPathSchema returns a schema for directory path parameters
func DirectoryPathSchema() Property {
	return StringProperty("Absolute or relative path to a directory").WithFormat("uri")
}

// GlobPatternSchema returns a schema for glob pattern parameters
func GlobPatternSchema() Property {
	return StringProperty("Glob pattern for matching files (e.g., '**/*.go')").WithPattern("^[^*]*[*][^*]*$")
}

// TimeoutSchema returns a schema for timeout parameters
func TimeoutSchema() Property {
	return IntegerProperty("Timeout in milliseconds").WithDefault(120000).WithMin(0)
}

// LimitSchema returns a schema for limit parameters
func LimitSchema() Property {
	return IntegerProperty("Maximum number of results to return").WithDefault(100).WithMin(1)
}

// OffsetSchema returns a schema for offset parameters
func OffsetSchema() Property {
	return IntegerProperty("Number of results to skip").WithDefault(0).WithMin(0)
}
