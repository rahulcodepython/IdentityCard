import "server-only"

import { apiFetch } from "@/lib/api/client"
import {
  type CreatePersonInput,
  createPersonSchema,
  importSummarySchema,
  peopleListResponseSchema,
  personResponseSchema,
  type UpdatePersonInput,
  updatePersonSchema,
} from "@/lib/validation/people"

export type PeopleFilter = { subEventId?: string; search?: string }

function toQueryString(filter: PeopleFilter) {
  const params = new URLSearchParams()
  if (filter.subEventId) params.set("sub_event_id", filter.subEventId)
  if (filter.search) params.set("search", filter.search)
  const qs = params.toString()
  return qs ? `?${qs}` : ""
}

export async function listPeople(eventId: string, filter: PeopleFilter = {}) {
  const data = await apiFetch(
    `/events/${eventId}/people${toQueryString(filter)}`
  )
  return peopleListResponseSchema.parse(data)
}

export async function getPerson(eventId: string, id: string) {
  const data = await apiFetch(`/events/${eventId}/people/${id}`)
  return personResponseSchema.parse(data)
}

export async function createPerson(eventId: string, input: CreatePersonInput) {
  const body = createPersonSchema.parse(input)
  const data = await apiFetch(`/events/${eventId}/people`, {
    method: "POST",
    body: JSON.stringify(body),
  })
  return personResponseSchema.parse(data)
}

export async function updatePerson(
  eventId: string,
  id: string,
  input: UpdatePersonInput
) {
  const body = updatePersonSchema.parse(input)
  const data = await apiFetch(`/events/${eventId}/people/${id}`, {
    method: "PATCH",
    body: JSON.stringify(body),
  })
  return personResponseSchema.parse(data)
}

export async function deletePerson(eventId: string, id: string) {
  await apiFetch(`/events/${eventId}/people/${id}`, { method: "DELETE" })
}

export async function importPeopleCsv(eventId: string, formData: FormData) {
  const data = await apiFetch(`/events/${eventId}/people/import`, {
    method: "POST",
    body: formData,
  })
  return importSummarySchema.parse(data)
}
