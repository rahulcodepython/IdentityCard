package generic

import (
    "time"
    "uuid"
)

// =====================================================================
// Common Wire Envelopes & Primitive Types
// =====================================================================

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

// MessageResponse provides a standard message payload.
type MessageResponse struct {
    Message string `json:"message"`
}

// UserID represents an application user identifier string.
type UserID string

// =====================================================================
// Organization Domain Types
// =====================================================================

type OrganizationDB struct {
    ID      uuid.UUID `json:"id"`
    Name    string    `json:"name"`
    Slug    string    `json:"slug"`
    HasLogo bool      `json:"has_logo"`
}

type DeleteOrganizationResult struct {
    UserOrgCount int  `json:"user_org_count"`
    Deleted      bool `json:"deleted"`
}

type UpdateOrganizationSettingsRequest struct {
    Name string `json:"name" validate:"required,min=1,max=100"`
}

type OrganizationSettingsResponse struct {
    ID      uuid.UUID `json:"id"`
    Name    string    `json:"name"`
    Slug    string    `json:"slug"`
    HasLogo bool      `json:"has_logo"`
}

type ListOrganizationsResponse struct {
    ID   uuid.UUID `json:"id"`
    Name string    `json:"name"`
    Slug string    `json:"slug"`
    Logo string    `json:"logo,omitempty"`
    Role string    `json:"role"`
}

// =====================================================================
// Plans and Billing Domain Types
// =====================================================================

type PlanDB struct {
    ID               uuid.UUID `json:"id"`
    Code             string    `json:"code"`
    Kind             string    `json:"kind"`
    BillingCycle     string    `json:"billing_cycle"`
    Name             string    `json:"name"`
    Amount           *int64    `json:"amount"`
    PerEventAmount   *int64    `json:"per_event_amount"`
    Currency         string    `json:"currency"`
    EventQuota       *int      `json:"event_quota"`
    NominalIncrement int64     `json:"nominal_increment"`
    CreatedAt        time.Time `json:"created_at"`
    UpdatedAt        time.Time `json:"updated_at"`
}

type BillingDB struct {
    ID             uuid.UUID  `json:"id"`
    OrganizationID uuid.UUID  `json:"organization_id"`
    PlanID         uuid.UUID  `json:"plan_id"`
    LineageRootID  *uuid.UUID `json:"lineage_root_id"`
    BillingNumber  int        `json:"billing_number"`
    PeriodStart    time.Time  `json:"period_start"`
    PeriodEnd      time.Time  `json:"period_end"`
    PruneDate      time.Time  `json:"prune_date"`
    Status         string     `json:"status"`
    Amount         int64      `json:"amount"`
    PaidAt         *time.Time `json:"paid_at"`
    TransactionID  *uuid.UUID `json:"transaction_id"`
    CreatedAt      time.Time  `json:"created_at"`
    UpdatedAt      time.Time  `json:"updated_at"`
}

type CreditDB struct {
    ID              uuid.UUID  `json:"id"`
    OrganizationID  uuid.UUID  `json:"organization_id"`
    BillingID       uuid.UUID  `json:"billing_id"`
    Type            string     `json:"type"`
    EventID         *uuid.UUID `json:"event_id"`
    IsRestricted    bool       `json:"is_restricted"`
    RestrictedSince *time.Time `json:"restricted_since"`
    CreatedAt       time.Time  `json:"created_at"`
}

type TransactionDB struct {
    ID             uuid.UUID `json:"id"`
    OrganizationID uuid.UUID `json:"organization_id"`
    PlanID         uuid.UUID `json:"plan_id"`
    Amount         int64     `json:"amount"`
    Currency       string    `json:"currency"`
    CreatedAt      time.Time `json:"created_at"`
}

type PlanResponse struct {
    ID               uuid.UUID `json:"id"`
    Code             string    `json:"code"`
    Kind             string    `json:"kind"`
    BillingCycle     string    `json:"billing_cycle"`
    Name             string    `json:"name"`
    Amount           *int64    `json:"amount,omitempty"`
    PerEventAmount   *int64    `json:"per_event_amount,omitempty"`
    Currency         string    `json:"currency"`
    EventQuota       *int      `json:"event_quota,omitempty"`
    NominalIncrement int64     `json:"nominal_increment"`
}

type BillingResponse struct {
    ID            uuid.UUID `json:"id"`
    LineageRootID uuid.UUID `json:"lineage_root_id"`
    PlanCode      string    `json:"plan_code"`
    Kind          string    `json:"kind"`
    BillingCycle  string    `json:"billing_cycle"`
    BillingNumber int       `json:"billing_number"`
    PeriodStart   string    `json:"period_start"`
    PeriodEnd     string    `json:"period_end"`
    PruneDate     string    `json:"prune_date,omitempty"`
    Status        string    `json:"status"`
    Amount        int64     `json:"amount"`
    Currency      string    `json:"currency"`
    PaidAt        *string   `json:"paid_at,omitempty"`
}

type CreditResponse struct {
    ID              uuid.UUID  `json:"id"`
    Type            string     `json:"type"`
    EventID         *uuid.UUID `json:"event_id,omitempty"`
    IsRestricted    bool       `json:"is_restricted"`
    RestrictedSince *string    `json:"restricted_since,omitempty"`
    CreatedAt       string     `json:"created_at"`
}

type TransactionResponse struct {
    ID        uuid.UUID `json:"id"`
    PlanCode  string    `json:"plan_code"`
    Amount    int64     `json:"amount"`
    Currency  string    `json:"currency"`
    CreatedAt string    `json:"created_at"`
}

type OrgBillingResponse struct {
    Billings        []BillingResponse `json:"billings"`
    Credits         []CreditResponse  `json:"credits"`
    AvailableByType map[string]int    `json:"available_by_type"`
}

type PurchaseRequest struct {
    PlanCode      string `json:"plan_code" validate:"required"`
    EventQuantity *int   `json:"event_quantity,omitempty" validate:"omitempty,min=1"`
}

type UpgradeRequest struct {
    PlanCode string `json:"plan_code" validate:"required"`
}

// =====================================================================
// Event Domain Types
// =====================================================================

type EventDB struct {
    ID             uuid.UUID  `json:"id"`
    OrganizationID uuid.UUID  `json:"organization_id"`
    EventType      string     `json:"event_type"`
    Status         string     `json:"status"`
    StartDate      time.Time  `json:"start_date"`
    EndDate        time.Time  `json:"end_date"`
    PublishedAt    *time.Time `json:"published_at"`
    CreatedAt      time.Time  `json:"created_at"`
    UpdatedAt      time.Time  `json:"updated_at"`
}

type EventDayDB struct {
    ID        uuid.UUID `json:"id"`
    EventID   uuid.UUID `json:"event_id"`
    Date      time.Time `json:"date"`
    EntryTime string    `json:"entry_time"`
    ExitTime  string    `json:"exit_time"`
    CreatedAt time.Time `json:"created_at"`
}

type EventMetadataDB struct {
    EventID                     uuid.UUID `json:"event_id"`
    Name                        string    `json:"name"`
    Venue                       *string   `json:"venue"`
    OrganizerName               *string   `json:"organizer_name"`
    ImageObjectKey              *string   `json:"image_object_key"`
    OrganizerSignatureObjectKey *string   `json:"organizer_signature_object_key"`
    CreatedAt                   time.Time `json:"created_at"`
    UpdatedAt                   time.Time `json:"updated_at"`
}

type EventDayInput struct {
    Date      string `json:"date" validate:"required,datetime=2006-01-02"`
    EntryTime string `json:"entry_time" validate:"required,datetime=15:04"`
    ExitTime  string `json:"exit_time" validate:"required,datetime=15:04"`
}

type EventDayResponse struct {
    Date      string `json:"date"`
    EntryTime string `json:"entry_time"`
    ExitTime  string `json:"exit_time"`
}

type CreateEventRequest struct {
    Name          string          `json:"name" validate:"required,min=2,max=200"`
    Venue         string          `json:"venue" validate:"omitempty,max=300"`
    OrganizerName string          `json:"organizer_name" validate:"omitempty,max=200"`
    StartDate     string          `json:"start_date" validate:"omitempty,datetime=2006-01-02"`
    EndDate       string          `json:"end_date" validate:"omitempty,datetime=2006-01-02"`
    EventType     string          `json:"event_type" validate:"omitempty"`
    Days          []EventDayInput `json:"days" validate:"omitempty,dive"`
    RangeStart    string          `json:"range_start" validate:"omitempty,datetime=2006-01-02"`
    RangeEnd      string          `json:"range_end" validate:"omitempty,datetime=2006-01-02"`
}

type UpdateEventRequest struct {
    Name          string          `json:"name" validate:"required,min=2,max=200"`
    Venue         string          `json:"venue" validate:"omitempty,max=300"`
    OrganizerName string          `json:"organizer_name" validate:"omitempty,max=200"`
    StartDate     string          `json:"start_date" validate:"omitempty,datetime=2006-01-02"`
    EndDate       string          `json:"end_date" validate:"omitempty,datetime=2006-01-02"`
    Days          []EventDayInput `json:"days" validate:"omitempty,dive"`
    RangeStart    string          `json:"range_start" validate:"omitempty,datetime=2006-01-02"`
    RangeEnd      string          `json:"range_end" validate:"omitempty,datetime=2006-01-02"`
}

type EventSummary struct {
    ID                    uuid.UUID `json:"id"`
    Name                  string    `json:"name"`
    EventType             string    `json:"event_type"`
    Status                string    `json:"status"`
    StartDate             string    `json:"start_date"`
    EndDate               string    `json:"end_date"`
    Venue                 *string   `json:"venue"`
    OrganizerName         *string   `json:"organizer_name"`
    HasImage              bool      `json:"has_image"`
    HasOrganizerSignature bool      `json:"has_organizer_signature"`
    PublishedAt           *string   `json:"published_at"`
}

type EventResponse struct {
    EventSummary
    Days []EventDayResponse `json:"days"`
}

type DayImportRowError struct {
    Row     int    `json:"row"`
    Message string `json:"message"`
}

type DayImportSummary struct {
    Imported int                 `json:"imported"`
    Skipped  int                 `json:"skipped"`
    Errors   []DayImportRowError `json:"errors"`
}

// =====================================================================
// SubEvent Domain Types
// =====================================================================

type CreateSubEventRequest struct {
    Name      string `json:"name" validate:"required,min=2,max=200"`
    Date      string `json:"date" validate:"required,datetime=2006-01-02"`
    EntryTime string `json:"entry_time" validate:"required,datetime=15:04"`
    ExitTime  string `json:"exit_time" validate:"required,datetime=15:04"`
}

type UpdateSubEventRequest struct {
    Name      string `json:"name" validate:"required,min=2,max=200"`
    Date      string `json:"date" validate:"required,datetime=2006-01-02"`
    EntryTime string `json:"entry_time" validate:"required,datetime=15:04"`
    ExitTime  string `json:"exit_time" validate:"required,datetime=15:04"`
}

type SubEventResponse struct {
    ID        uuid.UUID `json:"id"`
    Name      string    `json:"name"`
    Date      string    `json:"date"`
    EntryTime string    `json:"entry_time"`
    ExitTime  string    `json:"exit_time"`
}

// =====================================================================
// People Domain Types
// =====================================================================

type CreatePersonRequest struct {
    Email       string      `json:"email" validate:"required,email"`
    Mobile      string      `json:"mobile" validate:"required,min=3,max=32"`
    Name        string      `json:"name" validate:"required,min=1,max=200"`
    ImageURL    string      `json:"image_url" validate:"omitempty,url,max=2048"`
    Age         *int16      `json:"age" validate:"omitempty,min=0,max=150"`
    Gender      string      `json:"gender" validate:"omitempty,max=50"`
    SubEventIDs []uuid.UUID `json:"sub_event_ids" validate:"omitempty,dive"`
}

type UpdatePersonRequest struct {
    Name        string      `json:"name" validate:"required,min=1,max=200"`
    ImageURL    string      `json:"image_url" validate:"omitempty,url,max=2048"`
    Age         *int16      `json:"age" validate:"omitempty,min=0,max=150"`
    Gender      string      `json:"gender" validate:"omitempty,max=50"`
    SubEventIDs []uuid.UUID `json:"sub_event_ids" validate:"omitempty,dive"`
}

type PersonResponse struct {
    ID          uuid.UUID   `json:"id"`
    Email       string      `json:"email"`
    Mobile      string      `json:"mobile"`
    Name        string      `json:"name"`
    ImageURL    *string     `json:"image_url"`
    Age         *int16      `json:"age"`
    Gender      *string     `json:"gender"`
    JoinedAt    *string     `json:"joined_at"`
    CardSentAt  *string     `json:"card_sent_at"`
    SubEventIDs []uuid.UUID `json:"sub_event_ids"`
}

type PeopleImportRowError struct {
    Row     int    `json:"row"`
    Message string `json:"message"`
}

type PeopleImportSummary struct {
    Inserted int                    `json:"inserted"`
    Updated  int                    `json:"updated"`
    Skipped  int                    `json:"skipped"`
    Errors   []PeopleImportRowError `json:"errors"`
}

type PeopleSubmitInput struct {
    Email      string
    Mobile     string
    Name       string
    ImageURL   string
    Age        *int16
    Gender     string
    SubEventID *uuid.UUID
}

type PeopleListFilter struct {
    SubEventID *uuid.UUID
    Search     *string
}

type PeopleImportOptions struct {
    SubEventID *uuid.UUID
}

// =====================================================================
// Form Domain Types
// =====================================================================

type PublicFormDB struct {
    ID               uuid.UUID  `json:"id"`
    Token            string     `json:"token"`
    Capacity         *int32     `json:"capacity"`
    SubmissionsCount int32      `json:"submissions_count"`
    IsActive         bool       `json:"is_active"`
    EventID          uuid.UUID  `json:"event_id"`
    SubEventID       *uuid.UUID `json:"sub_event_id"`
    OrganizationID   uuid.UUID  `json:"organization_id"`
    EventName        string     `json:"event_name"`
    SubEventName     *string    `json:"sub_event_name"`
}

type CreateFormRequest struct {
    SubEventID *uuid.UUID `json:"sub_event_id" validate:"omitempty"`
    Capacity   *int32     `json:"capacity" validate:"omitempty,min=1"`
}

type UpdateFormRequest struct {
    Capacity *int32 `json:"capacity" validate:"omitempty,min=1"`
    IsActive bool   `json:"is_active"`
}

type FormResponse struct {
    ID               uuid.UUID  `json:"id"`
    Token            string     `json:"token"`
    SubEventID       *uuid.UUID `json:"sub_event_id"`
    Capacity         *int32     `json:"capacity"`
    SubmissionsCount int32      `json:"submissions_count"`
    IsActive         bool       `json:"is_active"`
}

type PublicFormResponse struct {
    EventName    string  `json:"event_name"`
    SubEventName *string `json:"sub_event_name"`
    IsOpen       bool    `json:"is_open"`
}

type SubmitFormRequest struct {
    Email    string `json:"email" validate:"required,email"`
    Mobile   string `json:"mobile" validate:"required,min=3,max=32"`
    Name     string `json:"name" validate:"required,min=1,max=200"`
    ImageURL string `json:"image_url" validate:"omitempty,url,max=2048"`
    Age      *int16 `json:"age" validate:"omitempty,min=0,max=150"`
    Gender   string `json:"gender" validate:"omitempty,max=50"`
}

// =====================================================================
// Device Domain Types
// =====================================================================

type DeviceDB struct {
    ID             uuid.UUID  `json:"id"`
    OrganizationID uuid.UUID  `json:"organization_id"`
    Name           string     `json:"name"`
    KeyHash        *string    `json:"key_hash"`
    OtpCode        *string    `json:"otp_code"`
    OtpExpiresAt   *time.Time `json:"otp_expires_at"`
    Status         string     `json:"status"`
    VerifiedAt     *time.Time `json:"verified_at"`
    CreatedAt      time.Time  `json:"created_at"`
    UpdatedAt      time.Time  `json:"updated_at"`
}

type CreateDeviceRequest struct {
    Name string `json:"name" validate:"required,min=1,max=100"`
}

type DeviceResponse struct {
    ID         uuid.UUID `json:"id"`
    Name       string    `json:"name"`
    Status     string    `json:"status"`
    CreatedAt  string    `json:"created_at"`
    VerifiedAt *string   `json:"verified_at"`
}

type CreateDeviceResponse struct {
    DeviceResponse
    OTPCode      string `json:"otp_code"`
    OTPExpiresAt string `json:"otp_expires_at"`
}

type PairDeviceRequest struct {
    OTPCode string `json:"otp_code" validate:"required,len=6,numeric"`
}

type PairDeviceResponse struct {
    DeviceID         uuid.UUID `json:"device_id"`
    OrganizationName string    `json:"organization_name"`
    Key              string    `json:"key"`
}

type DeviceMeResponse struct {
    DeviceID         uuid.UUID `json:"device_id"`
    DeviceName       string    `json:"device_name"`
    OrganizationName string    `json:"organization_name"`
}

type DeviceClaims struct {
    DeviceID       uuid.UUID
    OrganizationID uuid.UUID
}

// =====================================================================
// Attendance Domain Types
// =====================================================================

type AttendanceRecordDB struct {
    ID             uuid.UUID  `json:"id"`
    OrganizationID uuid.UUID  `json:"organization_id"`
    EventID        uuid.UUID  `json:"event_id"`
    PersonID       uuid.UUID  `json:"person_id"`
    Date           time.Time  `json:"date"`
    EntryAt        *time.Time `json:"entry_at"`
    EntryStatus    *string    `json:"entry_status"`
    EntryDeviceID  *uuid.UUID `json:"entry_device_id"`
    ExitAt         *time.Time `json:"exit_at"`
    ExitStatus     *string    `json:"exit_status"`
    ExitDeviceID   *uuid.UUID `json:"exit_device_id"`
    CreatedAt      time.Time  `json:"created_at"`
    UpdatedAt      time.Time  `json:"updated_at"`
}

type AttendanceRowDB struct {
    ID             uuid.UUID  `json:"id"`
    OrganizationID uuid.UUID  `json:"organization_id"`
    EventID        uuid.UUID  `json:"event_id"`
    PersonID       uuid.UUID  `json:"person_id"`
    Date           string     `json:"date"`
    EntryAt        *string    `json:"entry_at"`
    EntryStatus    *string    `json:"entry_status"`
    EntryDeviceID  *uuid.UUID `json:"entry_device_id"`
    ExitAt         *string    `json:"exit_at"`
    ExitStatus     *string    `json:"exit_status"`
    ExitDeviceID   *uuid.UUID `json:"exit_device_id"`
    PersonName     string     `json:"person_name"`
}

type ScanRequest struct {
    QRToken string `json:"qr_token" validate:"required"`
}

type ScanResponse struct {
    Direction string         `json:"direction"`
    Status    *string        `json:"status"`
    Date      string         `json:"date"`
    EventName string         `json:"event_name"`
    Person    PersonResponse `json:"person"`
}

type RosterEntry struct {
    PersonID    uuid.UUID `json:"person_id"`
    PersonName  string    `json:"person_name"`
    PersonEmail string    `json:"person_email"`
    Date        string    `json:"date"`
    Attended    bool      `json:"attended"`
    EntryAt     *string   `json:"entry_at"`
    EntryStatus *string   `json:"entry_status"`
    ExitAt      *string   `json:"exit_at"`
    ExitStatus  *string   `json:"exit_status"`
}

type RosterFilter struct {
    SubEventID *uuid.UUID
    Date       *string
    Attended   *bool
    Status     *string
}

// =====================================================================
// Analytics Domain Types
// =====================================================================

type StatusCounts struct {
    Early  int `json:"early"`
    OnTime int `json:"on_time"`
    Late   int `json:"late"`
}

type DailyBreakdown struct {
    Date          string       `json:"date"`
    Expected      int          `json:"expected"`
    Present       int          `json:"present"`
    Absent        int          `json:"absent"`
    EntryStatuses StatusCounts `json:"entry_statuses"`
    ExitStatuses  StatusCounts `json:"exit_statuses"`
}

type SubEventSummary struct {
    SubEventID   uuid.UUID `json:"sub_event_id"`
    SubEventName string    `json:"sub_event_name"`
    TotalPeople  int       `json:"total_people"`
    Attended     int       `json:"attended"`
    Absent       int       `json:"absent"`
}

type SummaryResponse struct {
    TotalPeople int               `json:"total_people"`
    Attended    int               `json:"attended"`
    Absent      int               `json:"absent"`
    SubEvents   []SubEventSummary `json:"sub_events"`
}

type DailyResponse struct {
    Days []DailyBreakdown `json:"days"`
}

type DailyTrendPoint struct {
    Date    string `json:"date"`
    Present int    `json:"present"`
    Absent  int    `json:"absent"`
}

type OverviewResponse struct {
    TotalEvents    int               `json:"total_events"`
    EventsByStatus map[string]int    `json:"events_by_status"`
    TotalPeople    int               `json:"total_people"`
    Attended       int               `json:"attended"`
    Absent         int               `json:"absent"`
    ActiveDevices  int               `json:"active_devices"`
    DailyTrend     []DailyTrendPoint `json:"daily_trend"`
}

// =====================================================================
// Card Domain Types
// =====================================================================

type SubEventSchedule struct {
    Name string
    Days []EventDayResponse
}

type CardRenderInput struct {
    OrgName            string
    OrgLogo            []byte
    Event              EventResponse
    Person             PersonResponse
    SubEvents          []SubEventSchedule
    PersonPhoto        []byte
    QRToken            string
    EventImage         []byte
    OrganizerSignature []byte
}

type ResendCardResponse struct {
    Message string `json:"message"`
}
