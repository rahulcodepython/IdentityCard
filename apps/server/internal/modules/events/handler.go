package events

import (
	"context"
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"identitycard-server/internal/httpx"
	"identitycard-server/internal/middleware"
)

// CardSender is implemented by cards.Service. Defined here, on the
// consumer side, rather than importing that package directly — cards
// itself depends on events.Service (to read event details), so importing
// it back here would cycle. main.go wires the concrete *cards.Service in,
// since it's the one place both packages are already imported.
type CardSender interface {
	SendForEvent(ctx context.Context, orgID, eventID uuid.UUID)
}

type Handler struct {
	service    *Service
	cardSender CardSender
}

func NewHandler(service *Service, cardSender CardSender) *Handler {
	return &Handler{service: service, cardSender: cardSender}
}

func (h *Handler) Create(c *fiber.Ctx) error {
	var req CreateEventRequest
	if err := httpx.BindAndValidate(c, &req); err != nil {
		return err
	}
	resp, err := h.service.Create(c.Context(), middleware.Claims(c).OrganizationID, req)
	if err != nil {
		return err
	}
	return httpx.OK(c, fiber.StatusCreated, resp)
}

func (h *Handler) List(c *fiber.Ctx) error {
	resp, err := h.service.List(c.Context(), middleware.Claims(c).OrganizationID)
	if err != nil {
		return err
	}
	return httpx.OK(c, fiber.StatusOK, resp)
}

func (h *Handler) Get(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httpx.NewError(fiber.StatusBadRequest, "bad_request", "invalid event id")
	}
	resp, err := h.service.Get(c.Context(), middleware.Claims(c).OrganizationID, id)
	if err != nil {
		return err
	}
	return httpx.OK(c, fiber.StatusOK, resp)
}

func (h *Handler) Update(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httpx.NewError(fiber.StatusBadRequest, "bad_request", "invalid event id")
	}
	var req UpdateEventRequest
	if err := httpx.BindAndValidate(c, &req); err != nil {
		return err
	}
	resp, err := h.service.Update(c.Context(), middleware.Claims(c).OrganizationID, id, req)
	if err != nil {
		return err
	}
	return httpx.OK(c, fiber.StatusOK, resp)
}

func (h *Handler) Publish(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httpx.NewError(fiber.StatusBadRequest, "bad_request", "invalid event id")
	}
	orgID := middleware.Claims(c).OrganizationID
	resp, err := h.service.Publish(c.Context(), orgID, id)
	if err != nil {
		return err
	}

	// Fire-and-forget: emailing every attendee their card shouldn't block
	// the publish response. This goroutine outlives the request (hence
	// context.Background(), not c.Context()) and MUST recover its own
	// panics — Fiber's recover middleware only guards the request
	// goroutine, not ones spawned from inside a handler; an unrecovered
	// panic here would crash the whole process.
	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("events: panic sending cards for event %s: %v", id, r)
			}
		}()
		h.cardSender.SendForEvent(context.Background(), orgID, id)
	}()

	return httpx.OK(c, fiber.StatusOK, resp)
}

func (h *Handler) Delete(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return httpx.NewError(fiber.StatusBadRequest, "bad_request", "invalid event id")
	}
	if err := h.service.Delete(c.Context(), middleware.Claims(c).OrganizationID, id); err != nil {
		return err
	}
	return c.SendStatus(fiber.StatusNoContent)
}
