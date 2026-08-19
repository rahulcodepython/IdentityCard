"use server"

import { ApiError } from "@/lib/api/client"
import { importDaysCsv, importExcludedDatesCsv } from "@/lib/api/events"
import type { DayImportSummary } from "@/lib/validation/events"

export type ImportActionResult = { summary: DayImportSummary } | { error: string }

export async function importDaysCsvAction(
  eventId: string,
  formData: FormData
): Promise<ImportActionResult> {
  try {
    const summary = await importDaysCsv(eventId, formData)
    return { summary }
  } catch (err) {
    if (err instanceof ApiError) return { error: err.message }
    return { error: "Something went wrong. Please try again." }
  }
}

export async function importExclusionsCsvAction(
  eventId: string,
  formData: FormData
): Promise<ImportActionResult> {
  try {
    const summary = await importExcludedDatesCsv(eventId, formData)
    return { summary }
  } catch (err) {
    if (err instanceof ApiError) return { error: err.message }
    return { error: "Something went wrong. Please try again." }
  }
}
