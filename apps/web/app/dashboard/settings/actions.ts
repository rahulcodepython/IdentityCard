"use server"

import { revalidatePath } from "next/cache"

import { ApiError } from "@/lib/api/client"
import { uploadOrgLogo } from "@/lib/api/organizations"

export type LogoActionResult = { error: string } | undefined

export async function uploadLogoAction(
  _prevState: LogoActionResult,
  formData: FormData
): Promise<LogoActionResult> {
  try {
    await uploadOrgLogo(formData)
  } catch (err) {
    if (err instanceof ApiError) {
      return { error: err.message }
    }
    return { error: "Something went wrong. Please try again." }
  }
  revalidatePath("/dashboard/settings")
}
