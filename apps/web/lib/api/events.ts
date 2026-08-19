import "server-only"

import { apiFetch } from "@/lib/api/client"
import {
  type AddExcludedDateInput,
  addExcludedDateSchema,
  type CreateEventInput,
  createEventSchema,
  dayImportSummarySchema,
  eventResponseSchema,
  eventsListResponseSchema,
  type UpdateEventInput,
  updateEventSchema,
} from "@/lib/validation/events"

export async function listEvents() {
  const data = await apiFetch("/events")
  return eventsListResponseSchema.parse(data)
}

export async function getEvent(id: string) {
  const data = await apiFetch(`/events/${id}`)
  return eventResponseSchema.parse(data)
}

export async function createEvent(input: CreateEventInput) {
  const body = createEventSchema.parse(input)
  const data = await apiFetch("/events", {
    method: "POST",
    body: JSON.stringify(body),
  })
  return eventResponseSchema.parse(data)
}

export async function updateEvent(id: string, input: UpdateEventInput) {
  const body = updateEventSchema.parse(input)
  const data = await apiFetch(`/events/${id}`, {
    method: "PATCH",
    body: JSON.stringify(body),
  })
  return eventResponseSchema.parse(data)
}

export async function publishEvent(id: string) {
  const data = await apiFetch(`/events/${id}/publish`, { method: "POST" })
  return eventResponseSchema.parse(data)
}

export async function deleteEvent(id: string) {
  await apiFetch(`/events/${id}`, { method: "DELETE" })
}

export async function addExcludedDate(id: string, input: AddExcludedDateInput) {
  const body = addExcludedDateSchema.parse(input)
  const data = await apiFetch(`/events/${id}/days/exclusions`, {
    method: "POST",
    body: JSON.stringify(body),
  })
  return eventResponseSchema.parse(data)
}

export async function importDaysCsv(id: string, formData: FormData) {
  const data = await apiFetch(`/events/${id}/days/import`, {
    method: "POST",
    body: formData,
  })
  return dayImportSummarySchema.parse(data)
}

export async function importExcludedDatesCsv(id: string, formData: FormData) {
  const data = await apiFetch(`/events/${id}/days/exclusions/import`, {
    method: "POST",
    body: formData,
  })
  return dayImportSummarySchema.parse(data)
}
