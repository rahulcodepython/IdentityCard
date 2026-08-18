package people

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"identitycard-server/internal/httpx"
	"identitycard-server/internal/middleware"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func parseEventID(c *fiber.Ctx) (uuid.UUID, error) {
	id, err := uuid.Parse(c.Params("eventId"))
	if err != nil {
		return uuid.UUID{}, httpx.NewError(fiber.StatusBadRequest, "bad_request", "invalid event id")
	}
	return id, nil
}

func (h *Handler) Create(c *fiber.Ctx) error {
	eventID, err := parseEventID(c)
	if err != nil {
		return err
	}
	var req CreatePersonRequest
	if err := httpx.BindAndValidate(c, &req); err != nil {
		return err
	}
	resp, err := h.service.Create(c.Context(), middleware.Claims(c).OrganizationID, eventID, req)
	if err != nil {
		return err
	}
	return httpx.OK(c, fiber.StatusCreated, resp)
}

func (h *Handler) List(c *fiber.Ctx) error {
	eventID, err := parseEventID(c)
	if err != nil {
		return err
	}
	resp, err := h.service.List(c.Context(), middleware.Claims(c).OrganizationID, eventID, filterFromQuery(c))
	if err != nil {
		return err
	}
	return httpx.OK(c, fiber.StatusOK, resp)
}

func (h *Handler) Get(c *fiber.Ctx) error {
	eventID, err := parseEventID(c)
	if err != nil {
		return err
	}
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httpx.NewError(fiber.StatusBadRequest, "bad_request", "invalid person id")
	}
	resp, err := h.service.Get(c.Context(), middleware.Claims(c).OrganizationID, eventID, id)
	if err != nil {
		return err
	}
	return httpx.OK(c, fiber.StatusOK, resp)
}

func (h *Handler) Update(c *fiber.Ctx) error {
	eventID, err := parseEventID(c)
	if err != nil {
		return err
	}
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httpx.NewError(fiber.StatusBadRequest, "bad_request", "invalid person id")
	}
	var req UpdatePersonRequest
	if err := httpx.BindAndValidate(c, &req); err != nil {
		return err
	}
	resp, err := h.service.Update(c.Context(), middleware.Claims(c).OrganizationID, eventID, id, req)
	if err != nil {
		return err
	}
	return httpx.OK(c, fiber.StatusOK, resp)
}

func (h *Handler) Delete(c *fiber.Ctx) error {
	eventID, err := parseEventID(c)
	if err != nil {
		return err
	}
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httpx.NewError(fiber.StatusBadRequest, "bad_request", "invalid person id")
	}
	if err := h.service.Delete(c.Context(), middleware.Claims(c).OrganizationID, eventID, id); err != nil {
		return err
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func (h *Handler) Import(c *fiber.Ctx) error {
	eventID, err := parseEventID(c)
	if err != nil {
		return err
	}

	fileHeader, err := c.FormFile("file")
	if err != nil {
		return httpx.NewError(fiber.StatusBadRequest, "bad_request", "missing CSV file")
	}
	file, err := fileHeader.Open()
	if err != nil {
		return httpx.ErrInternal()
	}
	defer file.Close()

	var opts ImportOptions
	if raw := c.FormValue("sub_event_id"); raw != "" {
		id, err := uuid.Parse(raw)
		if err != nil {
			return httpx.NewError(fiber.StatusBadRequest, "bad_request", "invalid sub_event_id")
		}
		opts.SubEventID = &id
	}

	summary, err := h.service.ImportCSV(c.Context(), middleware.Claims(c).OrganizationID, eventID, file, opts)
	if err != nil {
		return err
	}
	return httpx.OK(c, fiber.StatusOK, summary)
}

// Export returns a CSV file directly — the one handler in this module
// (in the whole API, alongside none other yet) that doesn't use the
// standard {"data": ...} envelope, since the response is a file download.
func (h *Handler) Export(c *fiber.Ctx) error {
	eventID, err := parseEventID(c)
	if err != nil {
		return err
	}

	var ids []uuid.UUID
	if raw := c.Query("ids"); raw != "" {
		for _, part := range strings.Split(raw, ",") {
			id, err := uuid.Parse(strings.TrimSpace(part))
			if err != nil {
				return httpx.NewError(fiber.StatusBadRequest, "bad_request", "invalid id in ids")
			}
			ids = append(ids, id)
		}
	}

	csvBytes, err := h.service.Export(c.Context(), middleware.Claims(c).OrganizationID, eventID, filterFromQuery(c), ids)
	if err != nil {
		return err
	}

	c.Set(fiber.HeaderContentType, "text/csv; charset=utf-8")
	c.Set(fiber.HeaderContentDisposition, `attachment; filename="people.csv"`)
	return c.Send(csvBytes)
}

func filterFromQuery(c *fiber.Ctx) ListFilter {
	var filter ListFilter
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
