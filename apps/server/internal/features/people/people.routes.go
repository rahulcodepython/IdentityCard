package people

import (
	"mime/multipart"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"identitycard-server/internal/generic"
	"identitycard-server/internal/middlewares"
	"identitycard-server/internal/utils"
)

func (a *App) RegisterRoutes(router fiber.Router) {
	manage := middlewares.RequireRole(generic.RoleAdmin, generic.RoleSuperAdmin)

	g := router.Group("/events/:eventId/people", middlewares.RequireAuth, middlewares.RequireOrganization, manage)
	g.Post("/", a.handleCreate)
	g.Get("/", a.handleList)
	g.Get("/export", a.handleExport)
	g.Post("/import", a.handleImport)
	g.Get("/:id", a.handleGet)
	g.Patch("/:id", a.handleUpdate)
	g.Delete("/:id", a.handleDelete)
}

func (a *App) handleCreate(c *fiber.Ctx) error {
	eventID, err := parsePeopleEventID(c)
	if err != nil {
		return err
	}
	var req CreatePersonRequest
	if err := utils.BindAndValidate(c, &req); err != nil {
		return err
	}
	resp, err := a.Create(c.Context(), middlewares.Claims(c).OrganizationID, eventID, req)
	if err != nil {
		return err
	}
	return utils.OK(c, fiber.StatusCreated, resp)
}

func (a *App) handleList(c *fiber.Ctx) error {
	eventID, err := parsePeopleEventID(c)
	if err != nil {
		return err
	}
	resp, err := a.List(c.Context(), middlewares.Claims(c).OrganizationID, eventID, peopleFilterFromQuery(c))
	if err != nil {
		return err
	}
	return utils.OK(c, fiber.StatusOK, resp)
}

func (a *App) handleGet(c *fiber.Ctx) error {
	eventID, err := parsePeopleEventID(c)
	if err != nil {
		return err
	}
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return utils.NewError(fiber.StatusBadRequest, "bad_request", "invalid person id")
	}
	resp, err := a.Get(c.Context(), middlewares.Claims(c).OrganizationID, eventID, id)
	if err != nil {
		return err
	}
	return utils.OK(c, fiber.StatusOK, resp)
}

func (a *App) handleUpdate(c *fiber.Ctx) error {
	eventID, err := parsePeopleEventID(c)
	if err != nil {
		return err
	}
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return utils.NewError(fiber.StatusBadRequest, "bad_request", "invalid person id")
	}
	var req UpdatePersonRequest
	if err := utils.BindAndValidate(c, &req); err != nil {
		return err
	}
	resp, err := a.Update(c.Context(), middlewares.Claims(c).OrganizationID, eventID, id, req)
	if err != nil {
		return err
	}
	return utils.OK(c, fiber.StatusOK, resp)
}

func (a *App) handleDelete(c *fiber.Ctx) error {
	eventID, err := parsePeopleEventID(c)
	if err != nil {
		return err
	}
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return utils.NewError(fiber.StatusBadRequest, "bad_request", "invalid person id")
	}
	if err := a.Delete(c.Context(), middlewares.Claims(c).OrganizationID, eventID, id); err != nil {
		return err
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func (a *App) handleImport(c *fiber.Ctx) error {
	eventID, err := parsePeopleEventID(c)
	if err != nil {
		return err
	}

	file, err := openUploadedCSV(c)
	if err != nil {
		return err
	}
	defer file.Close()

	var opts PeopleImportOptions
	if raw := c.FormValue("sub_event_id"); raw != "" {
		id, err := uuid.Parse(raw)
		if err != nil {
			return utils.NewError(fiber.StatusBadRequest, "bad_request", "invalid sub_event_id")
		}
		opts.SubEventID = &id
	}

	summary, err := a.ImportCSV(c.Context(), middlewares.Claims(c).OrganizationID, eventID, file, opts)
	if err != nil {
		return err
	}
	return utils.OK(c, fiber.StatusOK, summary)
}

func (a *App) handleExport(c *fiber.Ctx) error {
	eventID, err := parsePeopleEventID(c)
	if err != nil {
		return err
	}

	var ids []uuid.UUID
	if raw := c.Query("ids"); raw != "" {
		for _, part := range strings.Split(raw, ",") {
			id, err := uuid.Parse(strings.TrimSpace(part))
			if err != nil {
				return utils.NewError(fiber.StatusBadRequest, "bad_request", "invalid id in ids")
			}
			ids = append(ids, id)
		}
	}

	csvBytes, err := a.Export(c.Context(), middlewares.Claims(c).OrganizationID, eventID, peopleFilterFromQuery(c), ids)
	if err != nil {
		return err
	}

	c.Set(fiber.HeaderContentType, "text/csv; charset=utf-8")
	c.Set(fiber.HeaderContentDisposition, `attachment; filename="people.csv"`)
	return c.Send(csvBytes)
}

func parsePeopleEventID(c *fiber.Ctx) (uuid.UUID, error) {
	id, err := uuid.Parse(c.Params("eventId"))
	if err != nil {
		return uuid.UUID{}, utils.NewError(fiber.StatusBadRequest, "bad_request", "invalid event id")
	}
	return id, nil
}

func peopleFilterFromQuery(c *fiber.Ctx) PeopleListFilter {
	var filter PeopleListFilter
	if raw := c.Query("sub_event_id"); raw != "" {
		if id, err := uuid.Parse(raw); err == nil {
			filter.SubEventID = &id
		}
	}
	if search := c.Query("search"); search != "" {
		filter.Search = &search
	}
	return filter
}

func openUploadedCSV(c *fiber.Ctx) (multipart.File, error) {
	fileHeader, err := c.FormFile("file")
	if err != nil {
		return nil, utils.NewError(fiber.StatusBadRequest, "bad_request", "missing CSV file")
	}
	file, err := fileHeader.Open()
	if err != nil {
		return nil, utils.ErrInternal()
	}
	return file, nil
}
