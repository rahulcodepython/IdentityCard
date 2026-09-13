package generic

// =====================================================================
// API Route Prefixes & Routing Constants
// =====================================================================
const (
    APIV1Prefix = "/api/v1"
)

// =====================================================================
// Standard Error Codes
// =====================================================================
const (
    ErrCodeBadRequest      = "bad_request"
    ErrCodeNotFound        = "not_found"
    ErrCodeUnauthorized    = "unauthorized"
    ErrCodeForbidden       = "forbidden"
    ErrCodeConflict        = "conflict"
    ErrCodeInternal        = "internal_error"
    ErrCodePayloadTooLarge = "payload_too_large"
    ErrCodeUnprocessable   = "unprocessable_entity"
)

// =====================================================================
// Standard Error Messages
// =====================================================================
const (
    ErrMsgUnauthorized       = "missing or invalid authorization"
    ErrMsgForbidden          = "forbidden: insufficient permissions"
    ErrMsgInvalidRequest     = "invalid request body or query parameters"
    ErrMsgInvalidRequestBody = "Invalid request body."
    ErrMsgValidationFailed   = "Validation failed."
    ErrMsgTooManyRequests    = "too many requests, please try again later"
    ErrMsgInternal           = "internal server error"
)

// =====================================================================
// Standard Success Messages
// =====================================================================
const (
    MsgStatusOk = "ok"
)

// =====================================================================
// HTTP Header Constants
// =====================================================================
const (
    HeaderAuthorization      = "Authorization"
    HeaderContentType        = "Content-Type"
    HeaderContentDisposition = "Content-Disposition"
)

// =====================================================================
// Content-Type / MIME Constants
// =====================================================================
const (
    ContentTypeJSON = "application/json"
    ContentTypePDF  = "application/pdf"
    ContentTypeCSV  = "text/csv; charset=utf-8"
    ContentTypePNG  = "image/png"
    ContentTypeJPEG = "image/jpeg"
)

// =====================================================================
// Date and Time Layouts
// =====================================================================
const (
    DateFormat     = "2006-01-02"
    TimeFormat     = "15:04"
    DateTimeFormat = "2006-01-02 15:04:05"
)

// =====================================================================
// Pagination Defaults & Limits
// =====================================================================
const (
    DefaultPage  = 1
    DefaultLimit = 50
    MaxPage      = 1_000_000_000
)

