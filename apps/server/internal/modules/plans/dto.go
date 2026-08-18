package plans

import "github.com/google/uuid"

// PriceConfig is the shape stored in plans.price_config (jsonb). Phase 1
// only needs enough to display a price; the dynamic, storage-scaling
// pricing engine for the custom recurring tier is a later phase.
type PriceConfig struct {
	Amount   int64  `json:"amount"` // smallest currency unit (e.g. paise); 0/omitted for "contact sales"
	Currency string `json:"currency"`
	Billing  string `json:"billing"` // one_time | per_event | yearly | custom
}

type PlanResponse struct {
	ID    uuid.UUID   `json:"id"`
	Code  string      `json:"code"`
	Kind  string      `json:"kind"`
	Tier  string      `json:"tier"`
	Name  string      `json:"name"`
	Price PriceConfig `json:"price"`
}
