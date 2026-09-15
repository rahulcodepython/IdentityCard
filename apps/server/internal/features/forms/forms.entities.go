package forms

import (
    "time"
)

// SupportedFieldTypes list of allowable form field types.
var SupportedFieldTypes = map[string]bool{
    "text":     true,
    "email":    true,
    "number":   true,
    "url":      true,
    "checkbox": true,
    "radio":    true,
    "switch":   true,
    "date":     true,
    "time":     true,
    "month":    true,
    "week":     true,
    "file":     true,
}

// FieldOption represents a selectable option for radio, checkbox, etc.
type FieldOption struct {
    ID    string `json:"id"`
    Label string `json:"label"`
    Value string `json:"value"`
}

// FieldValidation represents constraints for input validation.
type FieldValidation struct {
    Min           *float64 `json:"min,omitempty"`
    Max           *float64 `json:"max,omitempty"`
    Pattern       *string  `json:"pattern,omitempty"`
    Accept        *string  `json:"accept,omitempty"`
    MaxFileSizeMB *int     `json:"max_file_size_mb,omitempty"`
    Multiple      *bool    `json:"multiple,omitempty"`
}

// FormField defines the structure of each individual form field stored in JSONB.
type FormField struct {
    ID          string           `json:"id"`
    Key         string           `json:"key"`
    Label       string           `json:"label"`
    Type        string           `json:"type"`
    Required    bool             `json:"required"`
    Placeholder string           `json:"placeholder"`
    IsSystem    bool             `json:"is_system"`
    Options     []FieldOption    `json:"options"`
    Validation  *FieldValidation `json:"validation,omitempty"`
}

// Form represents a form entity in the database.
type Form struct {
    ID        string      `json:"id"`
    Name      string      `json:"name"`
    Fields    []FormField `json:"fields"`
    CreatedAt time.Time   `json:"created_at"`
    UpdatedAt time.Time   `json:"updated_at"`
}

// CreateFormRequest represents payload for creating a new form (accepts only Name).
type CreateFormRequest struct {
    Name string `json:"name" validate:"required"`
}

// UpdateFormRequest represents payload for editing form metadata.
type UpdateFormRequest struct {
    Name *string `json:"name" validate:"required"`
}

// UpdateFormFieldsRequest represents payload for reordering/updating all fields.
type UpdateFormFieldsRequest struct {
    Fields []FormField `json:"fields" validate:"required"`
}
