import "server-only"

import { apiFetch } from "@/lib/api/client"
import {
  type CreateSubEventInput,
  createSubEventSchema,
  subEventResponseSchema,
  subEventsListResponseSchema,
  type UpdateSubEventInput,
  updateSubEventSchema,
} from "@/lib/validation/subevents"

export async function listSubEvents(eventId: string) {
  const data = await apiFetch(`/events/${eventId}/subevents`)
  return subEventsListResponseSchema.parse(data)
}

export async function getSubEvent(eventId: string, id: string) {
  const data = await apiFetch(`/events/${eventId}/subevents/${id}`)
  return subEventResponseSchema.parse(data)
}

export async function createSubEvent(
  eventId: string,
  input: CreateSubEventInput
) {
  const body = createSubEventSchema.parse(input)
  const data = await apiFetch(`/events/${eventId}/subevents`, {
    method: "POST",
    body: JSON.stringify(body),
  })
  return subEventResponseSchema.parse(data)
}

export async function updateSubEvent(
  eventId: string,
  id: string,
  input: UpdateSubEventInput
) {
  const body = updateSubEventSchema.parse(input)
  const data = await apiFetch(`/events/${eventId}/subevents/${id}`, {
    method: "PATCH",
    body: JSON.stringify(body),
  })
  return subEventResponseSchema.parse(data)
}

export async function deleteSubEvent(eventId: string, id: string) {
  await apiFetch(`/events/${eventId}/subevents/${id}`, { method: "DELETE" })
}
