import { z } from "zod"

// Mirrors apps/server/internal/modules/organizations/dto.go.

export const settingsResponseSchema = z.object({
  id: z.string().uuid(),
  name: z.string(),
  slug: z.string(),
  has_logo: z.boolean(),
})
export type OrgSettings = z.infer<typeof settingsResponseSchema>
