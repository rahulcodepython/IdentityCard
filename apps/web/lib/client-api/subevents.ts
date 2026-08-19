import { z } from "zod"

import { apiRequest } from "@/react-query/client"
import {
  type CreateSubEventInput,
  createSubEventSchema,
  subEventResponseSchema,
  subEventsListResponseSchema,
  type UpdateSubEventInput,
  updateSubEventSchema,
} from "@/schema/subevents.types"

export async function listSubEvents(eventId: string) {
  return apiRequest(
    { url: `/events/${eventId}/subevents`, method: "GET" },
    subEventsListResponseSchema
  )
}

export async function getSubEvent(eventId: string, id: string) {
  return apiRequest(
    { url: `/events/${eventId}/subevents/${id}`, method: "GET" },
    subEventResponseSchema
  )
}

export async function createSubEvent(
  eventId: string,
  input: CreateSubEventInput
) {
  const body = createSubEventSchema.parse(input)
  return apiRequest(
    { url: `/events/${eventId}/subevents`, method: "POST", data: body },
    subEventResponseSchema
  )
}

export async function updateSubEvent(
  eventId: string,
  id: string,
  input: UpdateSubEventInput
) {
  const body = updateSubEventSchema.parse(input)
  return apiRequest(
    { url: `/events/${eventId}/subevents/${id}`, method: "PATCH", data: body },
    subEventResponseSchema
  )
}

export async function deleteSubEvent(eventId: string, id: string) {
  await apiRequest(
    { url: `/events/${eventId}/subevents/${id}`, method: "DELETE" },
    z.unknown()
  )
}