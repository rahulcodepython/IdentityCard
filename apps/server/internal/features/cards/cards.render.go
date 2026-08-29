package cards

import (
	"bytes"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"

	"github.com/go-pdf/fpdf"
	qrcode "github.com/skip2/go-qrcode"

	"identitycard-server/internal/features/events"
	"identitycard-server/internal/features/people"
)

const (
	cardWidthMM  = 95.0
	cardHeightMM = 150.0
)

type subEventSchedule struct {
	Name string
	Days []events.EventDayResponse
}

type renderInput struct {
	OrgName            string
	OrgLogo            []byte
	Event              events.EventResponse
	Person             people.PersonResponse
	SubEvents          []subEventSchedule
	PersonPhoto        []byte
	QRToken            string
	EventImage         []byte
	OrganizerSignature []byte
}

func renderCardPDF(in renderInput) ([]byte, error) {
	pdf := fpdf.NewCustom(&fpdf.InitType{
		OrientationStr: "P",
		UnitStr:        "mm",
		Size:           fpdf.SizeType{Wd: cardWidthMM, Ht: cardHeightMM},
	})
	pdf.SetMargins(6, 6, 6)
	pdf.SetAutoPageBreak(false, 0)
	pdf.AddPage()

	headerX := 6.0
	if in.OrgLogo != nil && registerAndDrawImage(pdf, "logo", in.OrgLogo, 6, 6, 20, 0) {
		headerX = 30
	}
	pdf.SetXY(headerX, 8)
	pdf.SetFont("Helvetica", "B", 12)
	pdf.CellFormat(cardWidthMM-headerX-6, 6, in.OrgName, "", 2, "L", false, 0, "")

	pdf.SetX(6)
	pdf.Ln(3)
	pdf.SetFont("Helvetica", "B", 14)
	pdf.MultiCell(cardWidthMM-12, 6, in.Event.Name, "", "L", false)

	if in.Event.OrganizerName != nil && *in.Event.OrganizerName != "" {
		pdf.SetFont("Helvetica", "I", 9)
		pdf.MultiCell(cardWidthMM-12, 4.5, "Organizer: "+*in.Event.OrganizerName, "", "L", false)
	}

	if in.Event.Venue != nil && *in.Event.Venue != "" {
		pdf.SetFont("Helvetica", "", 9)
		pdf.MultiCell(cardWidthMM-12, 5, "Venue: "+*in.Event.Venue, "", "L", false)
	}

	pdf.SetFont("Helvetica", "", 9)
	dateRange := in.Event.StartDate
	if in.Event.EndDate != "" && in.Event.EndDate != in.Event.StartDate {
		dateRange += " to " + in.Event.EndDate
	}
	pdf.CellFormat(0, 5, "Dates: "+dateRange, "", 1, "L", false, 0, "")
	pdf.Ln(2)

	detailsY := pdf.GetY()
	detailsX := 6.0
	if in.PersonPhoto != nil && registerAndDrawImage(pdf, "photo", in.PersonPhoto, 6, detailsY, 25, 25) {
		detailsX = 34
	}
	pdf.SetXY(detailsX, detailsY)
	pdf.SetFont("Helvetica", "B", 11)
	pdf.CellFormat(cardWidthMM-detailsX-6, 5, in.Person.Name, "", 2, "L", false, 0, "")
	pdf.SetX(detailsX)
	pdf.SetFont("Helvetica", "", 8)
	pdf.CellFormat(cardWidthMM-detailsX-6, 4, in.Person.Email, "", 2, "L", false, 0, "")
	pdf.SetX(detailsX)
	pdf.CellFormat(cardWidthMM-detailsX-6, 4, in.Person.Mobile, "", 2, "L", false, 0, "")
	if in.Person.Age != nil || in.Person.Gender != nil {
		pdf.SetX(detailsX)
		pdf.CellFormat(cardWidthMM-detailsX-6, 4, personMeta(in.Person), "", 2, "L", false, 0, "")
	}

	pdf.SetXY(6, detailsY+27)
	pdf.SetFont("Helvetica", "B", 10)
	pdf.CellFormat(0, 6, "Permitted schedule", "", 1, "L", false, 0, "")
	pdf.SetFont("Helvetica", "", 8)
	renderSchedule(pdf, in)

	qrPNG, err := qrcode.Encode(in.QRToken, qrcode.Medium, 256)
	if err != nil {
		return nil, fmt.Errorf("cards: generate qr: %w", err)
	}
	qrSize := 32.0
	qrX := (cardWidthMM - qrSize) / 2
	qrY := cardHeightMM - qrSize - 10
	if !registerAndDrawImage(pdf, "qr", qrPNG, qrX, qrY, qrSize, qrSize) {
		return nil, fmt.Errorf("cards: could not embed generated qr code")
	}

	pdf.SetFont("Helvetica", "I", 6)
	pdf.SetXY(6, cardHeightMM-6)
	pdf.CellFormat(cardWidthMM-12, 3, "Valid only for the dates/times listed above.", "", 0, "C", false, 0, "")

	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, fmt.Errorf("cards: render pdf: %w", err)
	}
	return buf.Bytes(), nil
}

func renderSchedule(pdf *fpdf.Fpdf, in renderInput) {
	if len(in.SubEvents) == 0 {
		for _, d := range in.Event.Days {
			pdf.CellFormat(0, 4.5, fmt.Sprintf("%s: %s - %s", d.Date, d.EntryTime, d.ExitTime), "", 1, "L", false, 0, "")
		}
		return
	}
	for _, se := range in.SubEvents {
		pdf.SetFont("Helvetica", "B", 8)
		pdf.CellFormat(0, 4.5, se.Name, "", 1, "L", false, 0, "")
		pdf.SetFont("Helvetica", "", 8)
		days := se.Days
		if len(days) == 0 {
			days = in.Event.Days
		}
		for _, d := range days {
			pdf.CellFormat(0, 4.5, fmt.Sprintf("  %s: %s - %s", d.Date, d.EntryTime, d.ExitTime), "", 1, "L", false, 0, "")
		}
	}
}

func personMeta(p people.PersonResponse) string {
	meta := ""
	if p.Age != nil {
		meta = fmt.Sprintf("Age: %d", *p.Age)
	}
	if p.Gender != nil {
		if meta != "" {
			meta += "  "
		}
		meta += "Gender: " + *p.Gender
	}
	return meta
}

func registerAndDrawImage(pdf *fpdf.Fpdf, name string, data []byte, x, y, w, h float64) bool {
	_, format, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil || (format != "png" && format != "jpeg") {
		return false
	}
	imageType := "JPG"
	if format == "png" {
		imageType = "PNG"
	}

	options := fpdf.ImageOptions{ImageType: imageType}
	pdf.RegisterImageOptionsReader(name, options, bytes.NewReader(data))
	if pdf.Error() != nil {
		return false
	}
	pdf.ImageOptions(name, x, y, w, h, false, options, 0, "")
	return pdf.Error() == nil
}
