package generic

// API Route Prefixes
const (
	APIV1Prefix = "/api/v1"
)

type Role string

// Organization Member Roles
const (
	RoleAdmin  Role = "admin"
	RoleMember Role = "member"
)

// Standard Error Codes
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

// Standard Error Messages
const (
	ErrMsgInvalidEventID       = "invalid event id"
	ErrMsgInvalidSubEventID    = "invalid sub_event_id"
	ErrMsgInvalidPersonID      = "invalid person id"
	ErrMsgInvalidDeviceID      = "invalid device id"
	ErrMsgInvalidFormID        = "invalid form id"
	ErrMsgInvalidLineageRootID = "invalid lineage root id"
	ErrMsgInvalidIDInList      = "invalid id in list"
	ErrMsgMissingImageFile     = "missing image file"
	ErrMsgMissingSignatureFile = "missing signature file"
	ErrMsgMissingLogoFile      = "missing logo file"
	ErrMsgMissingCSVFile       = "missing CSV file"
	ErrMsgUnauthorized         = "missing or invalid bearer token"
	ErrMsgOrganizationRequired = "active organization required"
	ErrMsgForbidden            = "forbidden: insufficient permissions"
	ErrMsgDeviceAuthRequired   = "missing or invalid device credentials"
	ErrMsgInvalidRequest       = "invalid request body or query parameters"
)

// Standard Success Messages
const (
	MsgStatusOk   = "ok"
	MsgRegistered = "registered"
	MsgCardResent = "card resent"
)

// HTTP Header Constants
const (
	HeaderAuthorization      = "Authorization"
	HeaderContentType        = "Content-Type"
	HeaderContentDisposition = "Content-Disposition"
	HeaderDeviceKey          = "X-Device-Key"
	HeaderDeviceSecret       = "X-Device-Secret"
)

// Content-Type Constants
const (
	ContentTypePDF = "application/pdf"
	ContentTypeCSV = "text/csv; charset=utf-8"
)

// Context and Local State Keys
const (
	ContextKeyClaims       = "jwt_claims"
	ContextKeyDeviceClaims = "device_claims"
)
