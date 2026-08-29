package services

import (
	"testing"

	"identitycard-server/internal/entities"
)

func TestRenderCardPDF(t *testing.T) {
	venue := "Main Hall"
	endDate := "2026-03-02"
	age := int16(29)
	gender := "Female"

	pdfBytes, err := renderCardPDF(renderInput{
		OrgName: "Acme Events",
		Event: entities.EventResponse{
			EventSummary: entities.EventSummary{
				Name: "Annual Conference", StartDate: "2026-03-01", EndDate: endDate, Venue: &venue,
			},
			Days: []entities.EventDayResponse{
				{Date: "2026-03-01", EntryTime: "09:00", ExitTime: "18:00"},
				{Date: "2026-03-02", EntryTime: "09:00", ExitTime: "17:00"},
			},
		},
		Person: entities.PersonResponse{
			Name: "Jane Doe", Email: "jane@example.com", Mobile: "+1 555 0100",
			Age: &age, Gender: &gender,
		},
		SubEvents: []subEventSchedule{
			{Name: "Workshop Day", Days: []entities.EventDayResponse{{Date: "2026-03-01", EntryTime: "10:00", ExitTime: "12:00"}}},
		},
		QRToken: "test-token",
	})
	if err != nil {
		t.Fatalf("renderCardPDF: %v", err)
	}
	if len(pdfBytes) < 100 || string(pdfBytes[:4]) != "%PDF" {
		t.Fatalf("output does not look like a PDF (len=%d)", len(pdfBytes))
	}
	t.Logf("generated %d bytes", len(pdfBytes))
}
