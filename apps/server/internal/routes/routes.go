// Package routes builds every repository/service/controller in the API
// and registers their routes. NewRouter is the one place cross-domain
// interfaces get their concrete types (see EventsCardSender, satisfied by
// *services.CardsService) — mirrors how cmd/server/main.go used to be the
// only file that wired every domain together before this layout existed.
package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"

	"identitycard-server/internal/config"
	"identitycard-server/internal/controllers"
	dbgen "identitycard-server/internal/db/sqlc/generated"
	"identitycard-server/internal/pkg/mailer"
	"identitycard-server/internal/pkg/storage"
	"identitycard-server/internal/repositories"
	"identitycard-server/internal/services"
	"identitycard-server/internal/utils"
)

type Router struct {
	app *fiber.App
	cfg *config.Config

	Organizations *controllers.OrganizationsController
	Plans         *controllers.PlansController
	Events        *controllers.EventsController
	SubEvents     *controllers.SubEventsController
	People        *controllers.PeopleController
	Forms         *controllers.FormsController
	Cards         *controllers.CardsController
	Devices       *controllers.DevicesController
	Attendance    *controllers.AttendanceController
	Analytics     *controllers.AnalyticsController

	// Exported so cmd/server/main.go can hand them to internal/jobs, the only
	// consumer outside this package that still needs concrete service
	// types (the billing/recurrence background sweeps).
	PlansService  *services.PlansService
	EventsService *services.EventsService

	// devicesService isn't needed outside this package — only unexported
	// since SetUp (a separate method from NewRouter) needs it to mount the
	// scanner group.
	devicesService *services.DevicesService
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
	orgsRepo := repositories.NewOrganizationsRepository(queries)
	orgsService := services.NewOrganizationsService(orgsRepo, objectStore)
	orgsController := controllers.NewOrganizationsController(orgsService)

	plansRepo := repositories.NewPlansRepository(queries)
	billingRepo := repositories.NewBillingRepository(queries)
	creditsRepo := repositories.NewCreditsRepository(queries)
	txRepo := repositories.NewTransactionsRepository(queries)
	plansService := services.NewPlansService(plansRepo, billingRepo, creditsRepo, txRepo, pool, queries)
	plansController := controllers.NewPlansController(plansService)

	eventsRepo := repositories.NewEventsRepository(queries)
	eventsService := services.NewEventsService(eventsRepo, pool, plansService, objectStore)

	subEventsRepo := repositories.NewSubEventsRepository(queries)
	subEventsService := services.NewSubEventsService(subEventsRepo, eventsService)
	subEventsController := controllers.NewSubEventsController(subEventsService)

	peopleRepo := repositories.NewPeopleRepository(queries)
	peopleService := services.NewPeopleService(peopleRepo, eventsService, subEventsService, pool)
	peopleController := controllers.NewPeopleController(peopleService)

	formsRepo := repositories.NewFormsRepository(queries)
	formsService := services.NewFormsService(formsRepo, eventsService, subEventsService, peopleService)
	formsController := controllers.NewFormsController(formsService)

	cardsService := services.NewCardsService(cfg, eventsService, peopleService, subEventsService, orgsService, mail)
	cardsController := controllers.NewCardsController(cardsService)

	devicesRepo := repositories.NewDevicesRepository(queries)
	devicesService := services.NewDevicesService(devicesRepo, orgsService)
	devicesController := controllers.NewDevicesController(devicesService)

	attendanceRepo := repositories.NewAttendanceRepository(queries)
	attendanceService := services.NewAttendanceService(cfg, attendanceRepo, eventsService, peopleService, subEventsService)
	attendanceController := controllers.NewAttendanceController(attendanceService)

	analyticsService := services.NewAnalyticsService(subEventsService, attendanceService, eventsService, devicesService)
	analyticsController := controllers.NewAnalyticsController(analyticsService)

	// EventsController is built last: it fires cardsService on publish
	// (see controllers.EventsCardSender) and can only take the concrete
	// service once it exists, since events can't import cards directly
	// (cards depends on events — see services/cards.service.go's doc
	// comment). This is the one place both concrete types are in scope.
	eventsController := controllers.NewEventsController(eventsService, cardsService)

	return &Router{
		app: app, cfg: cfg,
		Organizations: orgsController, Plans: plansController,
		Events: eventsController, SubEvents: subEventsController, People: peopleController,
		Forms: formsController, Cards: cardsController, Devices: devicesController,
		Attendance: attendanceController, Analytics: analyticsController,
		PlansService: plansService, EventsService: eventsService,
		devicesService: devicesService,
	}
}

// SetUp registers this router's routes plus the endpoints that don't
// belong to any one domain (health check, API docs).
func (r *Router) SetUp() {
	r.app.Get("/health", func(c *fiber.Ctx) error {
		return utils.OK(c, fiber.StatusOK, fiber.Map{"status": "ok"})
	})
	registerDocsRoutes(r.app)

	registerOrganizationsRoutes(r.app, r.cfg, r.Organizations)
	registerPlansRoutes(r.app, r.cfg, r.Plans)
	registerEventsRoutes(r.app, r.cfg, r.Events)
	registerSubEventsRoutes(r.app, r.cfg, r.SubEvents)
	registerPeopleRoutes(r.app, r.cfg, r.People)
	registerFormsRoutes(r.app, r.cfg, r.Forms)
	registerFormsPublicRoutes(r.app, r.Forms)
	registerCardsRoutes(r.app, r.cfg, r.Cards)
	registerDevicesRoutes(r.app, r.cfg, r.Devices)
	registerDevicesPublicRoutes(r.app, r.Devices)
	registerAttendanceRoutes(r.app, r.cfg, r.Attendance)
	registerAnalyticsRoutes(r.app, r.cfg, r.Analytics)

	// devices.RegisterScannerRoutes and attendance.RegisterScanRoute mount
	// onto the SAME device-key-authenticated /scanner group — this is the
	// one place that needs both controllers, so it's the one place that
	// wires them together.
	scannerGroup := registerDevicesScannerRoutes(r.app, r.devicesService, r.Devices)
	registerAttendanceScanRoute(scannerGroup, r.Attendance)
}
