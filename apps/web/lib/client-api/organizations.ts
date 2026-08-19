import { z } from "zod"

import { apiRequest } from "@/react-query/client"
import {
  type CreateOrganizationInput,
  createOrganizationSchema,
  settingsResponseSchema,
} from "@/schema/organizations.types"

export async function getOrgSettings() {
  return apiRequest(
    { url: "/organizations/settings", method: "GET" },
    settingsResponseSchema
  )
}

export async function updateOrgSettings(input: { name: string }) {
  await apiRequest(
    { url: "/organizations/settings", method: "PATCH", data: input },
    z.unknown()
  )
}

export async function uploadOrgLogo(formData: FormData) {
  await apiRequest(
    { url: "/organizations/logo", method: "POST", data: formData },
    z.unknown()
  )
}

export async function deleteOrgLogo() {
  await apiRequest(
    { url: "/organizations/logo", method: "DELETE" },
    z.unknown()
  )
}

export async function deleteOrganization() {
  await apiRequest({ url: "/organizations", method: "DELETE" }, z.unknown())
}

// Finishes onboarding for a signed-in user who doesn't have an
// organization yet (a Google signup) — see /onboarding.
export async function createOrganization(input: CreateOrganizationInput) {
  const body = createOrganizationSchema.parse(input)
  return apiRequest(
    { url: "/auth/organization", method: "POST", data: body },
    z.unknown()
  )
}