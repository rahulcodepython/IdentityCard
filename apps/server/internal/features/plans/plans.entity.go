package plans

import (
    "time"

    "github.com/google/uuid"
)

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
