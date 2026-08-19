"use server"

import { redirect } from "next/navigation"

import { createOrganization } from "@/lib/api/organizations"
import { ApiError } from "@/lib/api/client"
import {
  type CreateOrganizationInput,
  createOrganizationSchema,
} from "@/lib/validation/organizations"

export type OnboardingActionResult = { error: string } | undefined

export async function completeOnboardingAction(
  input: CreateOrganizationInput
): Promise<OnboardingActionResult> {
  const parsed = createOrganizationSchema.safeParse(input)
  if (!parsed.success) {
    return { error: "Enter your organization's name." }
  }

  try {
    await createOrganization(parsed.data)
  } catch (err) {
    if (err instanceof ApiError) return { error: err.message }
    return { error: "Something went wrong. Please try again." }
  }

  redirect("/dashboard")
}
