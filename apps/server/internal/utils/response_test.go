package utils_test

import (
	"encoding/json"
	"io"
	"net/http/httptest"
	"testing"

	"identitycard-server/internal/utils"

	"github.com/gofiber/fiber/v2"
)

func TestResponseEnvelope(t *testing.T) {
	app := fiber.New(fiber.Config{
		ErrorHandler: utils.ErrorHandler,
	})

	app.Get("/ok", func(c *fiber.Ctx) error {
		return utils.OK(c, fiber.StatusOK, fiber.Map{"foo": "bar"})
	})

	app.Get("/error", func(c *fiber.Ctx) error {
		return utils.ErrNotFound("Item not found.", nil)
	})

	// Test OK endpoint
	req := httptest.NewRequest("GET", "/ok", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("test /ok failed: %v", err)
	}
	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	var okRes utils.Response
	if err := json.Unmarshal(body, &okRes); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if !okRes.Success || okRes.Message != "success" {
		t.Fatalf("unexpected response envelope: %+v", okRes)
	}

	// Test Error endpoint
	errReq := httptest.NewRequest("GET", "/error", nil)
	errResp, err := app.Test(errReq)
	if err != nil {
		t.Fatalf("test /error failed: %v", err)
	}
	if errResp.StatusCode != fiber.StatusNotFound {
		t.Fatalf("expected 404, got %d", errResp.StatusCode)
	}

	errBody, _ := io.ReadAll(errResp.Body)
	var errRes utils.Response
	if err := json.Unmarshal(errBody, &errRes); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if errRes.Success || errRes.Error != "Item not found." {
		t.Fatalf("unexpected error response envelope: %+v", errRes)
	}
}
