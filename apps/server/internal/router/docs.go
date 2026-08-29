package router

import "github.com/gofiber/fiber/v2"

func registerDocsRoutes(app *fiber.App) {
	app.Get("/docs/openapi.yaml", func(c *fiber.Ctx) error {
		c.Set(fiber.HeaderContentType, "application/yaml")
		return c.SendString(openapiStub)
	})
	app.Get("/docs", func(c *fiber.Ctx) error {
		c.Set(fiber.HeaderContentType, fiber.MIMETextHTMLCharsetUTF8)
		return c.SendString(scalarPage)
	})
}

const openapiStub = `openapi: 3.0.3
info:
  title: IdentityCard API
  version: "1.0"
  description: |
    Stub spec — wired up early per the project's architecture template.
    Covers /health as a template; the rest of the surface (organizations,
    plans, events, sub-events, people, forms, cards, devices, attendance,
    analytics) gets filled in incrementally. Auth is entirely better-auth's
    (see apps/web) — every route here that needs a session expects a
    bearer JWT from it, verified against its JWKS endpoint.
servers:
  - url: /
paths:
  /health:
    get:
      summary: Liveness check
      responses:
        "200":
          description: OK
          content:
            application/json:
              schema:
                type: object
                properties:
                  data:
                    type: object
                    properties:
                      status:
                        type: string
                        example: ok
components:
  securitySchemes:
    bearerAuth:
      type: http
      scheme: bearer
      bearerFormat: JWT
`

const scalarPage = `<!doctype html>
<html>
  <head>
    <title>IdentityCard API Reference</title>
    <meta charset="utf-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1" />
  </head>
  <body>
    <script id="api-reference" data-url="/docs/openapi.yaml"></script>
    <script src="https://cdn.jsdelivr.net/npm/@scalar/api-reference"></script>
  </body>
</html>
`
