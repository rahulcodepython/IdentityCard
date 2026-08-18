"use server"

import { ApiError } from "@/lib/api/client"
import { importPeopleCsv } from "@/lib/api/people"
import type { ImportSummary } from "@/lib/validation/people"

export type ImportActionResult = { summary: ImportSummary } | { error: string }

export async function importCsvAction(
  eventId: string,
  formData: FormData
): Promise<ImportActionResult> {
  try {
    const summary = await importPeopleCsv(eventId, formData)
    return { summary }
  } catch (err) {
    if (err instanceof ApiError) {
      return { error: err.message }
    }
    return { error: "Something went wrong. Please try again." }
  }
}
