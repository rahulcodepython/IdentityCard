import "server-only"

import { apiFetch } from "@/lib/api/client"
import { settingsResponseSchema } from "@/lib/validation/organizations"

export async function getOrgSettings() {
  const data = await apiFetch("/organizations/settings")
  return settingsResponseSchema.parse(data)
}

export async function uploadOrgLogo(formData: FormData) {
  await apiFetch("/organizations/logo", { method: "POST", body: formData })
}
