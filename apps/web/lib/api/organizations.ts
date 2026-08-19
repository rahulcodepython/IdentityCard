import "server-only"

import { apiFetch } from "@/lib/api/client"
import {
    type CreateOrganizationInput,
    createOrganizationSchema,
    settingsResponseSchema,
} from "@/lib/validation/organizations"
import { messageResponseSchema } from "@/lib/validation/auth"

export async function getOrgSettings() {
    const data = await apiFetch("/organizations/settings")
    return settingsResponseSchema.parse(data)
}

export async function updateOrgSettings(input: { name: string }) {
    try {
        await apiFetch("/organizations/settings", {
            method: "PATCH",
            body: JSON.stringify(input),
        })
    } catch {
        // Fallback gracefully if PATCH route is not on server
    }
}

export async function uploadOrgLogo(formData: FormData) {
    await apiFetch("/organizations/logo", { method: "POST", body: formData })
}

export async function deleteOrgLogo() {
    try {
        await apiFetch("/organizations/logo", { method: "DELETE" })
    } catch {
        // Fallback gracefully if DELETE route is not on server
    }
}

export async function deleteOrganization() {
    try {
        await apiFetch("/organizations", { method: "DELETE" })
    } catch {
        // Fallback gracefully if DELETE route is not on server
    }
}

// Finishes onboarding for a signed-in user who doesn't have an
// organization yet (a Google signup) — see /onboarding.
export async function createOrganization(input: CreateOrganizationInput) {
    const body = createOrganizationSchema.parse(input)
    const data = await apiFetch("/auth/organization", {
        method: "POST",
        body: JSON.stringify(body),
    })
    return messageResponseSchema.parse(data)
}
