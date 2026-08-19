"use server"

import { revalidatePath } from "next/cache"

import { ApiError } from "@/lib/api/client"
import { renewSubscription, subscribeToPlan } from "@/lib/api/plans"
import type { SubscribeInput } from "@/lib/validation/plans"

export type BillingActionResult = { error: string } | { success: true }

export async function subscribeAction(
  input: SubscribeInput
): Promise<BillingActionResult> {
  try {
    await subscribeToPlan(input)
  } catch (err) {
    return { error: err instanceof ApiError ? err.message : "Something went wrong." }
  }
  revalidatePath("/dashboard")
  revalidatePath("/dashboard/billing")
  return { success: true }
}

export async function renewAction(
  subscriptionId: string
): Promise<BillingActionResult> {
  try {
    await renewSubscription(subscriptionId)
  } catch (err) {
    return { error: err instanceof ApiError ? err.message : "Something went wrong." }
  }
  revalidatePath("/dashboard")
  revalidatePath("/dashboard/billing")
  return { success: true }
}
