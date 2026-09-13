package billing

import (
    "time"
    "uuid"
)

type OrganizationBillingDB struct {
    OrganizationID     uuid.UUID  `json:"organization_id"`
    CreditBalance      int        `json:"credit_balance"`
    AnnualFeeStatus    string     `json:"annual_fee_status"`
    CurrentPeriodStart *time.Time `json:"current_period_start"`
    CurrentPeriodEnd   *time.Time `json:"current_period_end"`
    CancelAtPeriodEnd  bool       `json:"cancel_at_period_end"`
    UpdatedAt          time.Time  `json:"updated_at"`
    CreatedAt          time.Time  `json:"created_at"`
}

type BillingTransactionDB struct {
    ID             uuid.UUID  `json:"id"`
    OrganizationID uuid.UUID  `json:"organization_id"`
    Type           string     `json:"type"`
    CreditsDelta   int        `json:"credits_delta"`
    Amount         int64      `json:"amount"`
    Currency       string     `json:"currency"`
    EventID        *uuid.UUID `json:"event_id"`
    Description    *string    `json:"description"`
    CreatedAt      time.Time  `json:"created_at"`
}

type PricingConfig struct {
    CreditUnitPrice     int64  `json:"credit_unit_price"`
    Currency            string `json:"currency"`
    AnnualRenewalAmount int64  `json:"annual_renewal_amount"`
}

type BillingOverviewResponse struct {
    CreditBalance      int                    `json:"credit_balance"`
    AnnualFeeStatus    string                 `json:"annual_fee_status"`
    CurrentPeriodStart *string                `json:"current_period_start"`
    CurrentPeriodEnd   *string                `json:"current_period_end"`
    EventCount         int                    `json:"event_count"`
    Pricing            PricingConfig          `json:"pricing"`
    Transactions       []BillingTransactionDB `json:"transactions"`
}

type PurchaseCreditsRequest struct {
    Quantity int `json:"quantity" validate:"required,min=1,max=1000"`
}
