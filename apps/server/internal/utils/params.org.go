package utils

import (
    "github.com/gofiber/fiber/v2"
    "uuid"

    "identitycard-server/internal/generic"
)

// ParseOrgID extracts and validates the :orgId URL parameter from the request.
func ParseOrgID(c *fiber.Ctx) (uuid.UUID, error) {
    param := c.Params("orgId")
    if param == "" {
        return uuid.Nil(), ErrBadRequest("organization id is required", nil)
    }
    orgID, err := uuid.Parse(param)
    if err != nil {
        return uuid.Nil(), ErrBadRequest(generic.ErrMsgOrganizationRequired, err)
    }
    return orgID, nil
}
