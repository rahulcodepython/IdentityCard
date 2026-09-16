package applicants

import "time"

type ApplicantItem struct {
    UserID    string                 `json:"user_id"`
    Name      string                 `json:"name"`
    Email     string                 `json:"email"`
    Data      map[string]interface{} `json:"data"`
    CreatedAt time.Time              `json:"created_at"`
}

type ApplicantFilter struct {
    Field string `json:"field"`
    Op    string `json:"op"`
    Value string `json:"value"`
}

type FormFieldItem struct {
    ID          string      `json:"id"`
    Key         string      `json:"key"`
    Label       string      `json:"label"`
    Type        string      `json:"type"`
    Required    bool        `json:"required"`
    Placeholder string      `json:"placeholder,omitempty"`
    Options     interface{} `json:"options,omitempty"`
}

type FormSummary struct {
    ID     string          `json:"id"`
    Name   string          `json:"name"`
    Fields []FormFieldItem `json:"fields"`
}

type ListApplicantsResponse struct {
    Data  []ApplicantItem `json:"data"`
    Total int             `json:"total"`
    Page  int             `json:"page"`
    Limit int             `json:"limit"`
    Form  *FormSummary    `json:"form,omitempty"`
}
