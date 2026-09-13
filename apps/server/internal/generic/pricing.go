package generic

// =====================================================================
// Central Pricing & Billing Configuration
// All amounts are in the smallest currency unit (e.g. paise: 149900 = ₹1,499.00).
// Modify pricing or currency values directly here.
// =====================================================================
const (
    // CreditUnitPrice is the cost per single event creation credit.
    CreditUnitPrice int64 = 149900

    // CreditCurrency is the default billing currency.
    CreditCurrency = "INR"

    // InitialSignupCredits is the number of free credits granted to every new organization.
    InitialSignupCredits = 1

    // FlatAnnualRenewalAmount is the flat annual platform retention fee charged if events exist.
    FlatAnnualRenewalAmount int64 = 299900

    // AnnualGracePeriodDays is the grace window (in days) after annual expiration.
    AnnualGracePeriodDays = 30
)

// Annual Fee Statuses
const (
    AnnualFeeStatusFree    = "free"
    AnnualFeeStatusActive  = "active"
    AnnualFeeStatusPastDue = "past_due"
)

// Billing Transaction Types
const (
    TxTypeCreditPurchase = "credit_purchase"
    TxTypeCreditConsumed = "credit_consumed"
    TxTypeAnnualRenewal  = "annual_renewal"
)
