"use server"

import { revalidatePath } from "next/cache"

import { ApiError } from "@/lib/api/client"
import {
    deleteOrganization,
    deleteOrgLogo,
    updateOrgSettings,
    uploadOrgLogo,
} from "@/lib/api/organizations"

export type OrgActionResult = { error?: string; success?: boolean }

export async function updateOrgMetadataAction(name: string): Promise<OrgActionResult> {
    if (!name.trim()) {
        return { error: "Organization name cannot be empty." }
    }
    try {
        await updateOrgSettings({ name: name.trim() })
        revalidatePath("/dashboard/settings")
        return { success: true }
    } catch (err) {
        if (err instanceof ApiError) {
            return { error: err.message }
        }
        return { error: "Failed to update organization metadata." }
    }
}

export async function uploadLogoAction(formData: FormData): Promise<OrgActionResult> {
    try {
        await uploadOrgLogo(formData)
        revalidatePath("/dashboard/settings")
        return { success: true }
    } catch (err) {
        if (err instanceof ApiError) {
            return { error: err.message }
        }
        return { error: "Failed to upload logo." }
    }
}

export async function deleteLogoAction(): Promise<OrgActionResult> {
    try {
        await deleteOrgLogo()
        revalidatePath("/dashboard/settings")
        return { success: true }
    } catch (err) {
        if (err instanceof ApiError) {
            return { error: err.message }
        }
        return { error: "Failed to delete logo." }
    }
}

export async function deleteOrgAction(): Promise<OrgActionResult> {
    try {
        await deleteOrganization()
        revalidatePath("/dashboard")
        return { success: true }
    } catch (err) {
        if (err instanceof ApiError) {
            return { error: err.message }
        }
        return { error: "Failed to delete organization." }
    }
}
