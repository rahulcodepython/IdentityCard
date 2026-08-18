-- Tracks whether/when this person's ID card PDF was last emailed — set by
-- cards.Service after a successful send (bulk on publish, or a manual
-- resend). NULL means never sent.
ALTER TABLE people ADD COLUMN card_sent_at TIMESTAMPTZ;
