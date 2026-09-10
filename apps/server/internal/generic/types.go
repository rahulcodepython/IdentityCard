package generic

// Response is the canonical wire envelope.
type Response[T interface{}] struct {
    Success bool   `json:"success"`
    Message string `json:"message"`
    Data    T      `json:"data,omitempty"`
    Error   string `json:"error,omitempty"`
}

// PaginatedResponse wraps standard offset-paginated listings.
type PaginatedResponse[T interface{}] struct {
    Data  T   `json:"data"`
    Total int `json:"total"`
    Page  int `json:"page"`
    Limit int `json:"limit"`
}

// DeleteResponse returns the identifier of a deleted record.
type DeleteResponse struct {
    ID string `json:"id"`
}

// SuccessResponse confirms generic boolean outcomes.
type SuccessResponse struct {
    Success bool `json:"success"`
}

const (
    ErrMsgInvalidRequestBody = "Invalid request body."
    ErrMsgValidationFailed   = "Validation failed."
)
