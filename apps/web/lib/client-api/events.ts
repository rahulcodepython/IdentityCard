import { z } from "zod"

import { apiRequest } from "@/react-query/client"
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
} from "@/schema/events.types"

export async function listEvents() {
  return apiRequest({ url: "/events", method: "GET" }, eventsListResponseSchema)
}

export async function getEvent(id: string) {
  return apiRequest({ url: `/events/${id}`, method: "GET" }, eventResponseSchema)
}

export async function createEvent(input: CreateEventInput) {
  const body = createEventSchema.parse(input)
  return apiRequest(
    { url: "/events", method: "POST", data: body },
    eventResponseSchema
  )
}

export async function updateEvent(id: string, input: UpdateEventInput) {
  const body = updateEventSchema.parse(input)
  return apiRequest(
    { url: `/events/${id}`, method: "PATCH", data: body },
    eventResponseSchema
  )
}

export async function publishEvent(id: string) {
  return apiRequest(
    { url: `/events/${id}/publish`, method: "POST" },
    eventResponseSchema
  )
}

export async function deleteEvent(id: string) {
  await apiRequest({ url: `/events/${id}`, method: "DELETE" }, z.unknown())
}

export async function addExcludedDate(id: string, input: AddExcludedDateInput) {
  const body = addExcludedDateSchema.parse(input)
  return apiRequest(
    { url: `/events/${id}/days/exclusions`, method: "POST", data: body },
    eventResponseSchema
  )
}

export async function importDaysCsv(id: string, formData: FormData) {
  return apiRequest(
    { url: `/events/${id}/days/import`, method: "POST", data: formData },
    dayImportSummarySchema
  )
}

export async function importExcludedDatesCsv(id: string, formData: FormData) {
  return apiRequest(
    {
      url: `/events/${id}/days/exclusions/import`,
      method: "POST",
      data: formData,
    },
    dayImportSummarySchema
  )
}