package router

import (
	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"

	"identitycard-server/internal/config"
	dbgen "identitycard-server/internal/db/sqlc/generated"
	"identitycard-server/internal/features/analytics"
	"identitycard-server/internal/features/attendance"
	"identitycard-server/internal/features/cards"
	"identitycard-server/internal/features/devices"
	"identitycard-server/internal/features/events"
	"identitycard-server/internal/features/forms"
	"identitycard-server/internal/features/organizations"
	"identitycard-server/internal/features/people"
	"identitycard-server/internal/features/plans"
	"identitycard-server/internal/features/subevents"
	"identitycard-server/internal/pkg/mailer"
	"identitycard-server/internal/pkg/storage"
	"identitycard-server/internal/utils"
)

type Router struct {
	app *fiber.App
	cfg *config.Config

	Organizations *organizations.App
	Plans         *plans.App
	Events        *events.App
	SubEvents     *subevents.App
	People        *people.App
	Forms         *forms.App
	Cards         *cards.App
	Devices       *devices.App
	Attendance    *attendance.App
	Analytics     *analytics.App
}

func NewRouter(
	app *fiber.App,
	cfg *config.Config,
	pool *pgxpool.Pool,
	rdb *redis.Client,
	objectStore *storage.Storage,
	mail *mailer.Mailer,
	queries *dbgen.Queries,
) *Router {
	orgsApp := organizations.New(queries, objectStore)
	plansApp := plans.New(pool, queries)
	eventsApp := events.New(pool, queries, plansApp, objectStore)
	subEventsApp := subevents.New(queries, eventsApp)
	peopleApp := people.New(pool, queries, eventsApp, subEventsApp)
	formsApp := forms.New(queries, eventsApp, subEventsApp, peopleApp)
	cardsApp := cards.New(cfg, eventsApp, peopleApp, subEventsApp, orgsApp, mail)
	devicesApp := devices.New(queries, orgsApp)
	attendanceApp := attendance.New(cfg, queries, eventsApp, peopleApp, subEventsApp)
	analyticsApp := analytics.New(eventsApp, subEventsApp, attendanceApp, devicesApp)

	eventsApp.SetCardSender(cardsApp)

	return &Router{
		app:           app,
		cfg:           cfg,
		Organizations: orgsApp,
		Plans:         plansApp,
		Events:        eventsApp,
		SubEvents:     subEventsApp,
		People:        peopleApp,
		Forms:         formsApp,
		Cards:         cardsApp,
		Devices:       devicesApp,
		Attendance:    attendanceApp,
		Analytics:     analyticsApp,
	}
}

func (r *Router) SetUp() {
	r.app.Get("/health", func(c *fiber.Ctx) error {
		return utils.OK(c, fiber.StatusOK, fiber.Map{"status": "ok"})
	})
	registerDocsRoutes(r.app)

	r.Organizations.RegisterRoutes(r.app)
	r.Plans.RegisterRoutes(r.app)
	r.Events.RegisterRoutes(r.app)
	r.SubEvents.RegisterRoutes(r.app)
	r.People.RegisterRoutes(r.app)
	r.Forms.RegisterRoutes(r.app)
	r.Forms.RegisterPublicRoutes(r.app)
	r.Cards.RegisterRoutes(r.app)
	r.Devices.RegisterRoutes(r.app)
	r.Devices.RegisterPublicRoutes(r.app)
	r.Attendance.RegisterRoutes(r.app)
	r.Analytics.RegisterRoutes(r.app)

	scannerGroup := r.Devices.RegisterScannerRoutes(r.app)
	r.Attendance.RegisterScanRoute(scannerGroup)
}
