package main

import (
    "context"
    "encoding/json"
    "flag"
    "fmt"
    "log"
    "time"

    "github.com/jackc/pgx/v5/pgxpool"

    "identitycard-server/internal/config"
    "identitycard-server/internal/pkg/postgres"
)

func main() {
    action := flag.String("action", "seed", "Action: seed, flush, regenerate")
    flag.Parse()

    cfg, err := config.Load()
    if err != nil {
        log.Fatalf("[seed] failed to load config: %v", err)
    }

    ctx := context.Background()
    pool, err := postgres.Connect(ctx, cfg)
    if err != nil {
        log.Fatalf("[seed] failed to connect to database: %v", err)
    }
    defer postgres.Close(pool)

    switch *action {
    case "flush":
        if err := flushDatabase(ctx, pool); err != nil {
            log.Fatalf("[seed] flush failed: %v", err)
        }
        fmt.Println("✓ Database successfully flushed.")
    case "regenerate", "regenarate":
        if err := flushDatabase(ctx, pool); err != nil {
            log.Fatalf("[seed] flush before regenerate failed: %v", err)
        }
        fmt.Println("✓ Database flushed.")
        if err := seedDatabase(ctx, pool); err != nil {
            log.Fatalf("[seed] seed failed: %v", err)
        }
        fmt.Println("✓ Database successfully regenerated and seeded.")
    case "seed", "up":
        if err := seedDatabase(ctx, pool); err != nil {
            log.Fatalf("[seed] seed failed: %v", err)
        }
        fmt.Println("✓ Database successfully seeded.")
    default:
        log.Fatalf("[seed] unknown action: %s. Use seed, flush, or regenerate", *action)
    }
}

func flushDatabase(ctx context.Context, pool *pgxpool.Pool) error {
    flushSQL := `
        DO $$ 
        BEGIN
            IF EXISTS (SELECT FROM information_schema.tables WHERE table_schema = 'public' AND table_name = 'events') THEN
                TRUNCATE TABLE 
                    event_applicants, 
                    applicants, 
                    event_forms, 
                    form_templates, 
                    event_dates, 
                    event_metadata, 
                    events 
                RESTART IDENTITY CASCADE;
            END IF;
        END $$;
    `
    _, err := pool.Exec(ctx, flushSQL)
    return err
}

func mustJSON(v any) []byte {
    b, err := json.Marshal(v)
    if err != nil {
        panic(err)
    }
    return b
}

type fieldOption struct {
    ID    string `json:"id"`
    Label string `json:"label"`
    Value string `json:"value"`
}

type validationRule struct {
    Min       *float64 `json:"min,omitempty"`
    Max       *float64 `json:"max,omitempty"`
    MinLength *int     `json:"min_length,omitempty"`
    MaxLength *int     `json:"max_length,omitempty"`
}

type formFieldSeed struct {
    ID          string          `json:"id"`
    Key         string          `json:"key"`
    Label       string          `json:"label"`
    Type        string          `json:"type"`
    Required    bool            `json:"required"`
    Placeholder string          `json:"placeholder,omitempty"`
    IsSystem    bool            `json:"is_system"`
    Options     []fieldOption   `json:"options,omitempty"`
    Validation  *validationRule `json:"validation,omitempty"`
}

func intPtr(i int) *int {
    return &i
}

func floatPtr(f float64) *float64 {
    return &f
}

func seedDatabase(ctx context.Context, pool *pgxpool.Pool) error {
    tx, err := pool.Begin(ctx)
    if err != nil {
        return fmt.Errorf("begin transaction: %w", err)
    }
    defer tx.Rollback(ctx)

    // -------------------------------------------------------------
    // 1. Seed Global Form Templates (forms)
    // -------------------------------------------------------------
    t1Fields := []formFieldSeed{
        {
            ID: "sys-name", Key: "name", Label: "Full Name", Type: "text", Required: true,
            Placeholder: "John Doe", IsSystem: true,
            Validation: &validationRule{MinLength: intPtr(2), MaxLength: intPtr(80)},
        },
        {
            ID: "sys-email", Key: "email", Label: "Email Address", Type: "email", Required: true,
            Placeholder: "john@company.com", IsSystem: true,
        },
        {
            ID: "fld-phone", Key: "phone", Label: "Phone Number", Type: "text", Required: false,
            Placeholder: "+1 (555) 234-5678",
        },
        {
            ID: "fld-org", Key: "organization", Label: "Organization / Company", Type: "text", Required: true,
            Placeholder: "Acme Cloud Corp",
            Validation: &validationRule{MinLength: intPtr(2), MaxLength: intPtr(100)},
        },
        {
            ID: "fld-title", Key: "job_title", Label: "Job Title", Type: "text", Required: false,
            Placeholder: "Staff Infrastructure Engineer",
        },
        {
            ID: "fld-exp-lvl", Key: "experience_level", Label: "Seniority / Experience Level", Type: "radio", Required: true,
            Options: []fieldOption{
                {ID: "exp-1", Label: "Junior (0-2 years)", Value: "junior"},
                {ID: "exp-2", Label: "Mid-Level (3-5 years)", Value: "mid"},
                {ID: "exp-3", Label: "Senior (5-8 years)", Value: "senior"},
                {ID: "exp-4", Label: "Lead / Staff / Principal (8+ years)", Value: "lead"},
            },
        },
        {
            ID: "fld-tracks", Key: "tracks", Label: "Attending Tracks", Type: "checkbox", Required: true,
            Options: []fieldOption{
                {ID: "trk-1", Label: "AI & Machine Learning", Value: "ai_ml"},
                {ID: "trk-2", Label: "Cloud Architecture & Kubernetes", Value: "cloud_k8s"},
                {ID: "trk-3", Label: "Developer Experience & Frontend", Value: "devex_frontend"},
                {ID: "trk-4", Label: "Security & Zero Trust", Value: "security"},
                {ID: "trk-5", Label: "Data Engineering & Analytics", Value: "data_eng"},
            },
        },
        {
            ID: "fld-years", Key: "experience_years", Label: "Years in Tech", Type: "number", Required: false,
            Placeholder: "5",
            Validation: &validationRule{Min: floatPtr(0), Max: floatPtr(40)},
        },
        {
            ID: "fld-diet", Key: "dietary_requirements", Label: "Catering & Dietary Requirements", Type: "radio", Required: false,
            Options: []fieldOption{
                {ID: "diet-none", Label: "No Preference / Standard", Value: "standard"},
                {ID: "diet-veg", Label: "Vegetarian", Value: "vegetarian"},
                {ID: "diet-vegan", Label: "Vegan", Value: "vegan"},
                {ID: "diet-halal", Label: "Halal", Value: "halal"},
                {ID: "diet-kosher", Label: "Kosher", Value: "kosher"},
                {ID: "diet-gluten", Label: "Gluten-Free", Value: "gluten_free"},
            },
        },
        {
            ID: "fld-tshirt", Key: "tshirt_size", Label: "Conference Swag T-Shirt Size", Type: "radio", Required: false,
            Options: []fieldOption{
                {ID: "ts-s", Label: "Small (S)", Value: "S"},
                {ID: "ts-m", Label: "Medium (M)", Value: "M"},
                {ID: "ts-l", Label: "Large (L)", Value: "L"},
                {ID: "ts-xl", Label: "Extra Large (XL)", Value: "XL"},
                {ID: "ts-2xl", Label: "2X Large (2XL)", Value: "2XL"},
            },
        },
        {
            ID: "fld-bio", Key: "bio", Label: "Brief Attendee Bio", Type: "textarea", Required: false,
            Placeholder: "Share what projects you are building or what you hope to learn at the summit...",
            Validation: &validationRule{MaxLength: intPtr(500)},
        },
        {
            ID: "fld-newsletter", Key: "newsletter_opt_in", Label: "Subscribe to Event Updates & Community Newsletter", Type: "switch", Required: false,
        },
    }

    t2Fields := []formFieldSeed{
        {
            ID: "sys-name", Key: "name", Label: "Full Name", Type: "text", Required: true,
            Placeholder: "Ada Lovelace", IsSystem: true,
        },
        {
            ID: "sys-email", Key: "email", Label: "Email Address", Type: "email", Required: true,
            Placeholder: "ada@hack.io", IsSystem: true,
        },
        {
            ID: "fld-discord", Key: "discord_handle", Label: "Discord Handle", Type: "text", Required: true,
            Placeholder: "builder#0001",
        },
        {
            ID: "fld-team-status", Key: "team_status", Label: "Participation Format", Type: "radio", Required: true,
            Options: []fieldOption{
                {ID: "tm-solo", Label: "Solo Hacker", Value: "solo"},
                {ID: "tm-find", Label: "Looking for Team Members", Value: "need_team"},
                {ID: "tm-formed", Label: "Already Have a Team", Value: "have_team"},
            },
        },
        {
            ID: "fld-team-name", Key: "team_name", Label: "Team Name", Type: "text", Required: false,
            Placeholder: "Quantum Leapers",
        },
        {
            ID: "fld-role", Key: "primary_skill", Label: "Primary Hackathon Role", Type: "radio", Required: true,
            Options: []fieldOption{
                {ID: "sk-fe", Label: "Frontend Engineer", Value: "frontend"},
                {ID: "sk-be", Label: "Backend Engineer", Value: "backend"},
                {ID: "sk-ai", Label: "AI / ML Specialist", Value: "ai_ml"},
                {ID: "sk-des", Label: "UI/UX & Product Designer", Value: "designer"},
                {ID: "sk-pm", Label: "Product Manager / Pitcher", Value: "pm"},
            },
        },
        {
            ID: "fld-stack", Key: "tech_stack", Label: "Core Tech Stack", Type: "checkbox", Required: true,
            Options: []fieldOption{
                {ID: "st-ts", Label: "TypeScript / Next.js", Value: "typescript"},
                {ID: "st-py", Label: "Python / PyTorch", Value: "python"},
                {ID: "st-go", Label: "Go / Rust", Value: "go_rust"},
                {ID: "st-w3", Label: "Solidity / Web3", Value: "web3"},
                {ID: "st-db", Label: "PostgreSQL / Vector DB", Value: "databases"},
            },
        },
        {
            ID: "fld-hack-count", Key: "hackathons_attended", Label: "Previous Hackathons Participated", Type: "number", Required: false,
            Validation: &validationRule{Min: floatPtr(0), Max: floatPtr(50)},
        },
        {
            ID: "fld-hardware", Key: "needs_hardware", Label: "Requires Hardware Kit (Raspberry Pi / IoT Sensors)", Type: "switch", Required: false,
        },
        {
            ID: "fld-pitch", Key: "project_pitch", Label: "Preliminary Project Concept (Optional)", Type: "textarea", Required: false,
            Placeholder: "Outline your initial idea or problem statement you want to solve...",
            Validation: &validationRule{MaxLength: intPtr(1000)},
        },
    }

    t3Fields := []formFieldSeed{
        {
            ID: "sys-name", Key: "name", Label: "Full Name", Type: "text", Required: true, IsSystem: true,
        },
        {
            ID: "sys-email", Key: "email", Label: "Work Email", Type: "email", Required: true, IsSystem: true,
        },
        {
            ID: "fld-comp", Key: "company", Label: "Company / Organization", Type: "text", Required: true,
        },
        {
            ID: "fld-exec-title", Key: "executive_title", Label: "Executive Designation", Type: "text", Required: true,
            Placeholder: "Chief Technology Officer / VP",
        },
        {
            ID: "fld-session", Key: "session_type", Label: "Speaking Session Format", Type: "radio", Required: true,
            Options: []fieldOption{
                {ID: "ses-key", Label: "Keynote Address (40 min)", Value: "keynote"},
                {ID: "ses-fire", Label: "Fireside Chat (30 min)", Value: "fireside"},
                {ID: "ses-panel", Label: "Panelist Discussion (45 min)", Value: "panel"},
                {ID: "ses-tech", Label: "Technical Deep Dive (30 min)", Value: "technical"},
            },
        },
        {
            ID: "fld-dinner", Key: "vip_dinner_attendance", Label: "Attending Private VIP Executive Dinner", Type: "switch", Required: false,
        },
        {
            ID: "fld-notes", Key: "special_accommodations", Label: "Special Stage or AV Requirements", Type: "textarea", Required: false,
        },
    }

    t4Fields := []formFieldSeed{
        {
            ID: "sys-name", Key: "name", Label: "Full Name", Type: "text", Required: true, IsSystem: true,
        },
        {
            ID: "sys-email", Key: "email", Label: "Email Address", Type: "email", Required: true, IsSystem: true,
        },
        {
            ID: "fld-rsvp", Key: "attendance_mode", Label: "Attendance Mode", Type: "radio", Required: true,
            Options: []fieldOption{
                {ID: "mod-inperson", Label: "In-Person (San Francisco)", Value: "in_person"},
                {ID: "mod-virtual", Label: "Virtual Livestream", Value: "virtual"},
            },
        },
        {
            ID: "fld-guests", Key: "guest_count", Label: "Bringing Guests (+1)", Type: "number", Required: false,
            Placeholder: "0",
            Validation: &validationRule{Min: floatPtr(0), Max: floatPtr(3)},
        },
    }

    template1ID := "11111111-1111-1111-1111-111111111111"
    template2ID := "22222222-2222-2222-2222-222222222222"
    template3ID := "33333333-3333-3333-3333-333333333333"
    template4ID := "44444444-4444-4444-4444-444444444444"

    formsToInsert := []struct {
        ID     string
        Name   string
        Fields []byte
    }{
        {template1ID, "Tech Conference Standard Registration", mustJSON(t1Fields)},
        {template2ID, "Hackathon Team & Hacker Pass", mustJSON(t2Fields)},
        {template3ID, "Executive VIP & Speaker Accreditation", mustJSON(t3Fields)},
        {template4ID, "Community Meetup Quick RSVP", mustJSON(t4Fields)},
    }

    for _, f := range formsToInsert {
        _, err := tx.Exec(ctx, `
            INSERT INTO form_templates (id, name, fields, created_at, updated_at)
            VALUES ($1, $2, $3, now(), now())
            ON CONFLICT (id) DO UPDATE SET
                name = EXCLUDED.name,
                fields = EXCLUDED.fields,
                updated_at = now();
        `, f.ID, f.Name, f.Fields)
        if err != nil {
            return fmt.Errorf("insert form_template %s: %w", f.Name, err)
        }
    }

    // -------------------------------------------------------------
    // 2. Seed Events, Event Metadata, and Event Dates
    // -------------------------------------------------------------
    event1ID := "7e00b724-b5d0-454f-9b63-7a7362049fed"
    event2ID := "8e11c835-c6e1-4650-a774-8b84731500fe"
    event3ID := "9f22d946-d7f2-4761-b885-9c95842611ff"
    event4ID := "a033ea57-e803-4872-c996-0da695372200"
    event5ID := "b144fb68-f914-4983-da07-1eb7a6483311"
    event6ID := "c2550c79-0a25-4a94-eb18-2fc8b7594422"

    eventsData := []struct {
        ID        string
        StartDate string
        EndDate   string
        Name      string
        Venue     string
        Organizer string
        Logo      string
    }{
        {
            ID:        event1ID,
            StartDate: "2026-10-15",
            EndDate:   "2026-10-17",
            Name:      "Silicon Valley AI & Cloud Expo 2026",
            Venue:     "Moscone Center, Exhibit Hall A & B, San Francisco, CA",
            Organizer: "Cloud Native & AI Alliance",
            Logo:      "https://images.unsplash.com/photo-1540575467063-178a50c2df87?w=300",
        },
        {
            ID:        event2ID,
            StartDate: "2026-10-24",
            EndDate:   "2026-10-25",
            Name:      "NextGen Climate Hackathon 2026",
            Venue:     "Tech Hub Innovation Campus, Austin, TX",
            Organizer: "CleanCode & GreenEarth Foundation",
            Logo:      "https://images.unsplash.com/photo-1504384308090-c894fdcc538d?w=300",
        },
        {
            ID:        event3ID,
            StartDate: "2026-11-12",
            EndDate:   "2026-11-12",
            Name:      "FinTech Executive Leadership Summit",
            Venue:     "The Pierre Hotel, Grand Ballroom, New York, NY",
            Organizer: "Global FinTech Leaders Circle",
            Logo:      "https://images.unsplash.com/photo-1511578314322-379afb476865?w=300",
        },
        {
            ID:        event4ID,
            StartDate: "2026-11-05",
            EndDate:   "2026-11-06",
            Name:      "Design Systems & Web Architecture Workshop",
            Venue:     "Pier 27 Waterfront Event Center, San Francisco, CA",
            Organizer: "Frontend Collective",
            Logo:      "https://images.unsplash.com/photo-1531403009284-440f080d1e12?w=300",
        },
        {
            ID:        event5ID,
            StartDate: "2026-12-01",
            EndDate:   "2026-12-01",
            Name:      "Indie Founders & Builders Breakfast",
            Venue:     "Foundry Cafe & Co-working, Seattle, WA",
            Organizer: "Seattle Bootstrap Club",
            Logo:      "https://images.unsplash.com/photo-1517245386807-bb43f82c33c4?w=300",
        },
        {
            ID:        event6ID,
            StartDate: "2026-01-10",
            EndDate:   "2026-01-11",
            Name:      "Cyber Defense Winter Bootcamp (Past Event)",
            Venue:     "Virtual Security Operations Center",
            Organizer: "CyberSec Defense League",
            Logo:      "https://images.unsplash.com/photo-1550751827-4bd374c3f58b?w=300",
        },
    }

    for _, ed := range eventsData {
        _, err := tx.Exec(ctx, `
            INSERT INTO events (id, start_date, end_date, created_at, updated_at)
            VALUES ($1, $2, $3, now(), now())
            ON CONFLICT (id) DO UPDATE SET
                start_date = EXCLUDED.start_date,
                end_date = EXCLUDED.end_date,
                updated_at = now();
        `, ed.ID, ed.StartDate, ed.EndDate)
        if err != nil {
            return fmt.Errorf("insert event %s: %w", ed.Name, err)
        }

        _, err = tx.Exec(ctx, `
            INSERT INTO event_metadata (id, name, venue, organizer, logo)
            VALUES ($1, $2, $3, $4, $5)
            ON CONFLICT (id) DO UPDATE SET
                name = EXCLUDED.name,
                venue = EXCLUDED.venue,
                organizer = EXCLUDED.organizer,
                logo = EXCLUDED.logo;
        `, ed.ID, ed.Name, ed.Venue, ed.Organizer, ed.Logo)
        if err != nil {
            return fmt.Errorf("insert event_metadata %s: %w", ed.Name, err)
        }
    }

    // Event Dates schedule
    datesData := []struct {
        EventID   string
        Date      string
        StartTime string
        EndTime   string
    }{
        {event1ID, "2026-10-15", "09:00:00", "17:30:00"},
        {event1ID, "2026-10-16", "09:00:00", "18:00:00"},
        {event1ID, "2026-10-17", "10:00:00", "16:00:00"},

        {event2ID, "2026-10-24", "08:00:00", "23:00:00"},
        {event2ID, "2026-10-25", "08:00:00", "20:00:00"},

        {event3ID, "2026-11-12", "08:30:00", "17:00:00"},

        {event4ID, "2026-11-05", "09:30:00", "16:30:00"},
        {event4ID, "2026-11-06", "09:30:00", "16:30:00"},

        {event5ID, "2026-12-01", "08:30:00", "11:00:00"},

        {event6ID, "2026-01-10", "09:00:00", "18:00:00"},
        {event6ID, "2026-01-11", "09:00:00", "17:00:00"},
    }

    for _, d := range datesData {
        _, err := tx.Exec(ctx, `
            INSERT INTO event_dates (event_id, date, start_time, end_time, created_at, updated_at)
            VALUES ($1, $2, $3, $4, now(), now())
            ON CONFLICT (event_id, date) DO UPDATE SET
                start_time = EXCLUDED.start_time,
                end_time = EXCLUDED.end_time,
                updated_at = now();
        `, d.EventID, d.Date, d.StartTime, d.EndTime)
        if err != nil {
            return fmt.Errorf("insert event_date: %w", err)
        }
    }

    // -------------------------------------------------------------
    // 3. Seed Event Forms (Snapshot & Locking Architecture)
    // -------------------------------------------------------------
    // Custom scratch fields for Event 4
    e4ScratchFields := []formFieldSeed{
        {
            ID: "sys-name", Key: "name", Label: "Full Name", Type: "text", Required: true, IsSystem: true,
        },
        {
            ID: "sys-email", Key: "email", Label: "Email Address", Type: "email", Required: true, IsSystem: true,
        },
        {
            ID: "fld-role", Key: "current_role", Label: "Current Discipline", Type: "radio", Required: true,
            Options: []fieldOption{
                {ID: "r-ui", Label: "Product / UI Designer", Value: "designer"},
                {ID: "r-fe", Label: "Frontend Engineer", Value: "frontend_dev"},
                {ID: "r-dt", Label: "Design Technologist", Value: "design_technologist"},
                {ID: "r-ot", Label: "Other / Fullstack", Value: "other"},
            },
        },
        {
            ID: "fld-tool", Key: "primary_tool", Label: "Primary Design Tool", Type: "radio", Required: true,
            Options: []fieldOption{
                {ID: "t-figma", Label: "Figma", Value: "figma"},
                {ID: "t-penpot", Label: "Penpot", Value: "penpot"},
                {ID: "t-code", Label: "Code / Storybook Directly", Value: "code"},
            },
        },
        {
            ID: "fld-topics", Key: "topics_interest", Label: "Topics of Highest Interest", Type: "checkbox", Required: true,
            Options: []fieldOption{
                {ID: "top-tokens", Label: "Design Tokens & Cross-Platform Theming", Value: "tokens"},
                {ID: "top-a11y", Label: "Accessibility & WCAG 2.2 AAA Compliance", Value: "a11y"},
                {ID: "top-perf", Label: "Web Performance & CSS Architectures", Value: "perf"},
                {ID: "top-motion", Label: "Micro-Interactions & Motion Design", Value: "motion"},
            },
        },
        {
            ID: "fld-questions", Key: "questions_for_speakers", Label: "Questions for Workshop Instructors", Type: "textarea", Required: false,
            Placeholder: "What challenges is your team currently experiencing with your design system?",
        },
    }

    eventFormsToInsert := []struct {
        ID            string
        EventID       string
        Name          string
        Fields        []byte
        MaxApplicants int
        ExpiresAt     time.Time
        IsLocked      bool
        LockedAt      *time.Time
        Status        string
    }{
        // Event 1: Locked & Live, 100 max applicants, expires in future
        {
            ID:            "a1111111-aaaa-1111-aaaa-111111111111",
            EventID:       event1ID,
            Name:          "Silicon Valley AI Expo Registration",
            Fields:        mustJSON(t1Fields),
            MaxApplicants: 100,
            ExpiresAt:     time.Now().Add(30 * 24 * time.Hour),
            IsLocked:      true,
            LockedAt:      &[]time.Time{time.Now().Add(-5 * 24 * time.Hour)}[0],
            Status:        "live",
        },
        // Event 2: Locked & Live, 50 max applicants, expires in 2 weeks
        {
            ID:            "b2222222-bbbb-2222-bbbb-222222222222",
            EventID:       event2ID,
            Name:          "Climate Hackathon Hacker Pass Application",
            Fields:        mustJSON(t2Fields),
            MaxApplicants: 50,
            ExpiresAt:     time.Now().Add(14 * 24 * time.Hour),
            IsLocked:      true,
            LockedAt:      &[]time.Time{time.Now().Add(-3 * 24 * time.Hour)}[0],
            Status:        "live",
        },
        // Event 3: Unlocked Draft, 0 applicants (Organizers can edit or delete)
        {
            ID:            "c3333333-cccc-3333-cccc-333333333333",
            EventID:       event3ID,
            Name:          "Executive Speaker Nomination Form",
            Fields:        mustJSON(t3Fields),
            MaxApplicants: 60,
            ExpiresAt:     time.Now().Add(45 * 24 * time.Hour),
            IsLocked:      false,
            LockedAt:      nil,
            Status:        "waiting",
        },
        // Event 4: Locked & Live (Built from scratch)
        {
            ID:            "d4444444-dddd-4444-dddd-444444444444",
            EventID:       event4ID,
            Name:          "Design Systems Workshop Registration",
            Fields:        mustJSON(e4ScratchFields),
            MaxApplicants: 120,
            ExpiresAt:     time.Now().Add(25 * 24 * time.Hour),
            IsLocked:      true,
            LockedAt:      &[]time.Time{time.Now().Add(-2 * 24 * time.Hour)}[0],
            Status:        "live",
        },
        // Event 6: Locked & Expired
        {
            ID:            "f6666666-ffff-6666-ffff-666666666666",
            EventID:       event6ID,
            Name:          "Cyber Defense Winter Bootcamp Form",
            Fields:        mustJSON(t1Fields),
            MaxApplicants: 30,
            ExpiresAt:     time.Now().Add(-10 * 24 * time.Hour), // Expired!
            IsLocked:      true,
            LockedAt:      &[]time.Time{time.Now().Add(-60 * 24 * time.Hour)}[0],
            Status:        "live",
        },
    }

    for _, ef := range eventFormsToInsert {
        _, err := tx.Exec(ctx, `
            INSERT INTO event_forms (
                id, event_id, name, fields, 
                max_applicants, expires_at, status, is_locked, locked_at, created_at, updated_at
            )
            VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, now(), now())
            ON CONFLICT (event_id) DO UPDATE SET
                name = EXCLUDED.name,
                fields = EXCLUDED.fields,
                max_applicants = EXCLUDED.max_applicants,
                expires_at = EXCLUDED.expires_at,
                status = EXCLUDED.status,
                is_locked = EXCLUDED.is_locked,
                locked_at = EXCLUDED.locked_at,
                updated_at = now();
        `, ef.ID, ef.EventID, ef.Name, ef.Fields, ef.MaxApplicants, ef.ExpiresAt, ef.Status, ef.IsLocked, ef.LockedAt)
        if err != nil {
            return fmt.Errorf("insert event_form for event %s: %w", ef.EventID, err)
        }
    }

    // -------------------------------------------------------------
    // 4. Seed Diverse Applicants and Event Registrations
    // -------------------------------------------------------------
    type applicantData struct {
        ID        string
        Name      string
        Email     string
        Data      map[string]any
        EventID   string
        CreatedAt time.Time
    }

    now := time.Now()

    applicants := []applicantData{
        // Event 1 Applicants (AI & Cloud Expo)
        {
            ID: "USR-AI-01", Name: "Sarah Chen", Email: "sarah.chen@google.com", EventID: event1ID,
            CreatedAt: now.Add(-4 * 24 * time.Hour),
            Data: map[string]any{
                "organization": "Google DeepMind", "job_title": "Principal Research Scientist",
                "experience_level": "lead", "tracks": []string{"ai_ml", "cloud_k8s"},
                "experience_years": 12, "dietary_requirements": "vegetarian",
                "tshirt_size": "M", "newsletter_opt_in": true,
                "bio": "Researching autonomous reasoning systems and distributed model inference engines.",
            },
        },
        {
            ID: "USR-AI-02", Name: "Michael Scott", Email: "michael.scott@dunder.com", EventID: event1ID,
            CreatedAt: now.Add(-3 * 24 * time.Hour),
            Data: map[string]any{
                "organization": "Dunder Mifflin Tech", "job_title": "Regional Director of IT",
                "experience_level": "senior", "tracks": []string{"cloud_k8s", "devex_frontend"},
                "experience_years": 8, "dietary_requirements": "standard",
                "tshirt_size": "L", "newsletter_opt_in": false,
                "bio": "Managing multi-region hybrid cloud deployments across legacy paper inventory systems.",
            },
        },
        {
            ID: "USR-AI-03", Name: "Aisha Patel", Email: "aisha.patel@stripe.com", EventID: event1ID,
            CreatedAt: now.Add(-3 * 24 * time.Hour),
            Data: map[string]any{
                "organization": "Stripe", "job_title": "Staff Infrastructure Engineer",
                "experience_level": "lead", "tracks": []string{"cloud_k8s", "security"},
                "experience_years": 10, "dietary_requirements": "halal",
                "tshirt_size": "S", "newsletter_opt_in": true,
                "bio": "Architecting zero-downtime ledger consistency across distributed SQL clusters.",
            },
        },
        {
            ID: "USR-AI-04", Name: "Lucas Silva", Email: "lucas.silva@nubank.com", EventID: event1ID,
            CreatedAt: now.Add(-2 * 24 * time.Hour),
            Data: map[string]any{
                "organization": "Nubank", "job_title": "Software Engineer II",
                "experience_level": "mid", "tracks": []string{"devex_frontend", "data_eng"},
                "experience_years": 4, "dietary_requirements": "standard",
                "tshirt_size": "XL", "newsletter_opt_in": true,
            },
        },
        {
            ID: "USR-AI-05", Name: "Emily Watson", Email: "emily.watson@anthropic.com", EventID: event1ID,
            CreatedAt: now.Add(-2 * 24 * time.Hour),
            Data: map[string]any{
                "organization": "Anthropic", "job_title": "Safety & Alignment Researcher",
                "experience_level": "senior", "tracks": []string{"ai_ml", "security"},
                "experience_years": 7, "dietary_requirements": "vegan",
                "tshirt_size": "S", "newsletter_opt_in": true,
                "bio": "Investigating constitutional AI principles and mechanistic interpretability.",
            },
        },
        {
            ID: "USR-AI-06", Name: "Kenji Sato", Email: "kenji.sato@sony.co.jp", EventID: event1ID,
            CreatedAt: now.Add(-1 * 24 * time.Hour),
            Data: map[string]any{
                "organization": "Sony Interactive", "job_title": "Cloud Platform Architect",
                "experience_level": "lead", "tracks": []string{"cloud_k8s", "security", "data_eng"},
                "experience_years": 15, "dietary_requirements": "standard",
                "tshirt_size": "L", "newsletter_opt_in": false,
            },
        },
        {
            ID: "USR-AI-07", Name: "Liam O'Connor", Email: "liam.oconnor@shopify.com", EventID: event1ID,
            CreatedAt: now.Add(-18 * time.Hour),
            Data: map[string]any{
                "organization": "Shopify", "job_title": "Senior Frontend Developer",
                "experience_level": "senior", "tracks": []string{"devex_frontend"},
                "experience_years": 6, "dietary_requirements": "gluten_free",
                "tshirt_size": "M", "newsletter_opt_in": true,
            },
        },
        {
            ID: "USR-AI-08", Name: "Fatima Al-Mansoor", Email: "fatima.m@aramco.com", EventID: event1ID,
            CreatedAt: now.Add(-12 * time.Hour),
            Data: map[string]any{
                "organization": "Aramco Digital", "job_title": "Data Engineering Lead",
                "experience_level": "lead", "tracks": []string{"ai_ml", "data_eng"},
                "experience_years": 9, "dietary_requirements": "halal",
                "tshirt_size": "M", "newsletter_opt_in": true,
            },
        },
        {
            ID: "USR-AI-09", Name: "Chloe Martin", Email: "chloe.martin@datadoghq.com", EventID: event1ID,
            CreatedAt: now.Add(-6 * time.Hour),
            Data: map[string]any{
                "organization": "Datadog", "job_title": "Site Reliability Engineer",
                "experience_level": "mid", "tracks": []string{"cloud_k8s", "security"},
                "experience_years": 3, "dietary_requirements": "standard",
                "tshirt_size": "S", "newsletter_opt_in": false,
            },
        },
        {
            ID: "USR-AI-10", Name: "David Kim", Email: "david.kim@vercel.com", EventID: event1ID,
            CreatedAt: now.Add(-2 * time.Hour),
            Data: map[string]any{
                "organization": "Vercel", "job_title": "Developer Experience Engineer",
                "experience_level": "senior", "tracks": []string{"devex_frontend", "ai_ml"},
                "experience_years": 5, "dietary_requirements": "standard",
                "tshirt_size": "L", "newsletter_opt_in": true,
            },
        },

        // Event 2 Applicants (NextGen Climate Hackathon)
        {
            ID: "USR-HK-01", Name: "Julian Rivera", Email: "julian@ecohack.org", EventID: event2ID,
            CreatedAt: now.Add(-3 * 24 * time.Hour),
            Data: map[string]any{
                "discord_handle": "julian_rivera#4412", "team_status": "have_team",
                "team_name": "CarbonZero Squad", "primary_skill": "backend",
                "tech_stack": []string{"go_rust", "databases"},
                "hackathons_attended": 5, "needs_hardware": true,
                "project_pitch": "Building an open-source IoT grid monitor that calculates real-time marginal carbon intensity for EV charging.",
            },
        },
        {
            ID: "USR-HK-02", Name: "Maya Lin", Email: "maya.lin@mit.edu", EventID: event2ID,
            CreatedAt: now.Add(-2 * 24 * time.Hour),
            Data: map[string]any{
                "discord_handle": "mayacodes#1337", "team_status": "have_team",
                "team_name": "CarbonZero Squad", "primary_skill": "ai_ml",
                "tech_stack": []string{"python", "databases"},
                "hackathons_attended": 3, "needs_hardware": false,
                "project_pitch": "Satellite imagery analysis for automated detection of methane pipeline micro-leaks.",
            },
        },
        {
            ID: "USR-HK-03", Name: "Carlos Mendoza", Email: "carlos.m@utexas.edu", EventID: event2ID,
            CreatedAt: now.Add(-2 * 24 * time.Hour),
            Data: map[string]any{
                "discord_handle": "carlos_atx#9901", "team_status": "need_team",
                "primary_skill": "frontend",
                "tech_stack": []string{"typescript"},
                "hackathons_attended": 1, "needs_hardware": false,
                "project_pitch": "Interactive web dashboards visualizing groundwater depletion metrics.",
            },
        },
        {
            ID: "USR-HK-04", Name: "Hannah Schmidt", Email: "hannah.schmidt@berlin-tech.de", EventID: event2ID,
            CreatedAt: now.Add(-1 * 24 * time.Hour),
            Data: map[string]any{
                "discord_handle": "hannah_s#5522", "team_status": "solo",
                "primary_skill": "designer",
                "tech_stack": []string{"typescript"},
                "hackathons_attended": 4, "needs_hardware": true,
            },
        },
        {
            ID: "USR-HK-05", Name: "Tariq Mansour", Email: "tariq.mansour@gatech.edu", EventID: event2ID,
            CreatedAt: now.Add(-8 * time.Hour),
            Data: map[string]any{
                "discord_handle": "tariq_m#7711", "team_status": "have_team",
                "team_name": "SunGrid Optimization", "primary_skill": "backend",
                "tech_stack": []string{"go_rust", "python"},
                "hackathons_attended": 6, "needs_hardware": true,
            },
        },
        {
            ID: "USR-HK-06", Name: "Zoe Kravitz", Email: "zoe.kravitz@stanford.edu", EventID: event2ID,
            CreatedAt: now.Add(-3 * time.Hour),
            Data: map[string]any{
                "discord_handle": "zkravitz#0042", "team_status": "need_team",
                "primary_skill": "ai_ml",
                "tech_stack": []string{"python", "databases"},
                "hackathons_attended": 2, "needs_hardware": false,
            },
        },

        // Event 4 Applicants (Design Systems Workshop)
        {
            ID: "USR-DS-01", Name: "Oliver Queen", Email: "oliver.queen@airbnb.com", EventID: event4ID,
            CreatedAt: now.Add(-2 * 24 * time.Hour),
            Data: map[string]any{
                "current_role": "design_technologist", "primary_tool": "figma",
                "topics_interest": []string{"tokens", "perf"},
                "questions_for_speakers": "How do you coordinate design token deprecations across multi-platform iOS, Android, and Web teams?",
            },
        },
        {
            ID: "USR-DS-02", Name: "Jessica Jones", Email: "jessica.jones@netflix.com", EventID: event4ID,
            CreatedAt: now.Add(-1 * 24 * time.Hour),
            Data: map[string]any{
                "current_role": "frontend_dev", "primary_tool": "code",
                "topics_interest": []string{"a11y", "tokens"},
                "questions_for_speakers": "Best practices for automated CI regression testing on high-contrast accessibility themes.",
            },
        },
        {
            ID: "USR-DS-03", Name: "Arthur Curry", Email: "arthur.curry@atlantis.design", EventID: event4ID,
            CreatedAt: now.Add(-10 * time.Hour),
            Data: map[string]any{
                "current_role": "designer", "primary_tool": "figma",
                "topics_interest": []string{"tokens", "motion"},
            },
        },
        {
            ID: "USR-DS-04", Name: "Barry Allen", Email: "barry.allen@star.labs", EventID: event4ID,
            CreatedAt: now.Add(-4 * time.Hour),
            Data: map[string]any{
                "current_role": "frontend_dev", "primary_tool": "code",
                "topics_interest": []string{"perf", "motion"},
                "questions_for_speakers": "Minimizing bundle impact when shipping complex micro-interaction animations.",
            },
        },

        // Event 6 Applicants (Past Event)
        {
            ID: "USR-CY-01", Name: "Bruce Wayne", Email: "bruce@wayne-enterprises.com", EventID: event6ID,
            CreatedAt: now.Add(-50 * 24 * time.Hour),
            Data: map[string]any{
                "organization": "Wayne Enterprises Cyber", "job_title": "Chief Security Architect",
                "experience_level": "lead", "tracks": []string{"security"},
                "experience_years": 20, "dietary_requirements": "standard",
            },
        },
        {
            ID: "USR-CY-02", Name: "Clark Kent", Email: "clark.kent@dailyplanet.com", EventID: event6ID,
            CreatedAt: now.Add(-48 * 24 * time.Hour),
            Data: map[string]any{
                "organization": "Daily Planet Media", "job_title": "Investigative Tech Reporter",
                "experience_level": "senior", "tracks": []string{"security"},
                "experience_years": 10, "dietary_requirements": "vegetarian",
            },
        },
    }

    for _, a := range applicants {
        dataBytes := mustJSON(a.Data)

        // Insert into applicants master record
        _, err := tx.Exec(ctx, `
            INSERT INTO applicants (id, name, email, data, created_at, updated_at)
            VALUES ($1, $2, $3, $4, $5, now())
            ON CONFLICT (id) DO UPDATE SET
                name = EXCLUDED.name,
                email = EXCLUDED.email,
                data = EXCLUDED.data,
                updated_at = now();
        `, a.ID, a.Name, a.Email, dataBytes, a.CreatedAt)
        if err != nil {
            return fmt.Errorf("insert applicant %s: %w", a.Name, err)
        }

        // Insert into event_applicants join record
        _, err = tx.Exec(ctx, `
            INSERT INTO event_applicants (event_id, user_id, email, created_at)
            VALUES ($1, $2, $3, $4)
            ON CONFLICT (event_id, email) DO UPDATE SET
                user_id = EXCLUDED.user_id,
                created_at = EXCLUDED.created_at;
        `, a.EventID, a.ID, a.Email, a.CreatedAt)
        if err != nil {
            return fmt.Errorf("insert event_applicant for %s (%s): %w", a.Name, a.EventID, err)
        }
    }

    if err := tx.Commit(ctx); err != nil {
        return fmt.Errorf("commit transaction: %w", err)
    }

    return nil
}
