package generic

import "errors"

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

// =====================================================================
// Layer 2 — Per-Feature Domain Sentinel Catalog (DATA_LAYER_BLUEPRINT.md)
// =====================================================================

// Organizations errors
var (
    ErrOrgsNotFound     = errors.New("organization not found")
    ErrOrgsAccessDenied = errors.New("access denied for organization")
    ErrOrgsLogoNotFound = errors.New("organization logo not found")
)

// Plans and Billing errors
var (
    ErrPlansNotFound       = errors.New("plan not found")
    ErrPlansNoCredits      = errors.New("no credits available to create event")
    ErrPlansBillingNotFound = errors.New("billing record not found")
    ErrPlansAlreadyActive  = errors.New("plan is already active")
)

// Events errors
var (
    ErrEventsNotFound         = errors.New("event not found")
    ErrEventsAlreadyPublished = errors.New("event is already published")
    ErrEventsNotDraft         = errors.New("event must be in draft status to modify")
    ErrEventsInvalidDates     = errors.New("invalid event date range")
    ErrEventsDayHasAttendance = errors.New("cannot remove day with existing attendance records")
    ErrEventsAccessDenied     = errors.New("access denied for event")
)

// SubEvents errors
var (
    ErrSubEventsNotFound     = errors.New("sub-event not found")
    ErrSubEventsDateMismatch = errors.New("sub-event date must fall within parent event dates")
    ErrSubEventsAccessDenied = errors.New("access denied for sub-event")
)

// People errors
var (
    ErrPeopleNotFound     = errors.New("person not found")
    ErrPeopleDuplicate    = errors.New("person with this email or mobile already exists in event")
    ErrPeopleAccessDenied = errors.New("access denied for person")
    ErrPeopleCardNotSent  = errors.New("failed to send ID card")
)

// Forms errors
var (
    ErrFormsNotFound         = errors.New("form not found")
    ErrFormsInactive         = errors.New("form is no longer active")
    ErrFormsCapacityExceeded = errors.New("form registration capacity has been reached")
    ErrFormsAccessDenied     = errors.New("access denied for form")
)

// Devices errors
var (
    ErrDevicesNotFound       = errors.New("device not found")
    ErrDevicesInvalidOTP     = errors.New("invalid or expired device OTP")
    ErrDevicesAlreadyPaired  = errors.New("device is already paired")
    ErrDevicesRevoked        = errors.New("device has been revoked")
    ErrDevicesAccessDenied   = errors.New("access denied for device")
)

// Attendance errors
var (
    ErrAttendanceNotFound     = errors.New("attendance record not found")
    ErrAttendanceAlreadyEntry = errors.New("entry attendance already recorded for today")
    ErrAttendanceAlreadyExit  = errors.New("exit attendance already recorded for today")
    ErrAttendanceNoEntry      = errors.New("cannot record exit without prior entry")
    ErrAttendanceInvalidQR    = errors.New("invalid or expired QR token")
)
