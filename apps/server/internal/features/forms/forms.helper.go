package forms

import (
    "encoding/json"
)

type rawFieldCheck struct {
    Key      string `json:"key"`
    Type     string `json:"type"`
    IsSystem bool   `json:"is_system"`
}

// DefaultFormFields returns the minimal mandatory system fields required for every form template.
func DefaultFormFields() []FormField {
    return []FormField{
        {
            ID:          "field_default_name",
            Key:         "name",
            Label:       "Full Name",
            Type:        "text",
            Required:    true,
            Placeholder: "Enter your full name",
            IsSystem:    true,
            Options:     []FieldOption{},
        },
        {
            ID:          "field_default_email",
            Key:         "email",
            Label:       "Email Address",
            Type:        "email",
            Required:    true,
            Placeholder: "name@example.com",
            IsSystem:    true,
            Options:     []FieldOption{},
        },
    }
}

// defaultEventFormFields returns the raw JSON representation of default event form fields.
func defaultEventFormFields() []byte {
    return []byte(`[
        {"id":"field_default_name","key":"name","type":"text","label":"Full Name","options":[],"required":true,"is_system":true,"placeholder":"Enter your full name"},
        {"id":"field_default_email","key":"email","type":"email","label":"Email Address","options":[],"required":true,"is_system":true,"placeholder":"name@example.com"}
    ]`)
}

// validateEventFieldsJSON validates that the provided JSON contains mandatory name and email fields.
func validateEventFieldsJSON(raw []byte) error {
    if len(raw) == 0 {
        return ErrMissingMandatoryFields
    }
    var fields []rawFieldCheck
    if err := json.Unmarshal(raw, &fields); err != nil {
        return err
    }
    var hasName, hasEmail bool
    for _, f := range fields {
        if f.Key == "name" || (f.IsSystem && f.Type == "text") {
            hasName = true
        }
        if f.Key == "email" || (f.IsSystem && f.Type == "email") {
            hasEmail = true
        }
    }
    if !hasName || !hasEmail {
        return ErrMissingMandatoryFields
    }
    return nil
}
