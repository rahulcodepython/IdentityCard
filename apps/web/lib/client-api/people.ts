import { z } from "zod"

import { apiRequest } from "@/react-query/client"
import {
  type CreatePersonInput,
  createPersonSchema,
  importSummarySchema,
  peopleListResponseSchema,
  personResponseSchema,
  type UpdatePersonInput,
  updatePersonSchema,
} from "@/schema/people.types"

export type PeopleFilter = { subEventId?: string; search?: string }

function toQueryString(filter: PeopleFilter) {
  const params = new URLSearchParams()
  if (filter.subEventId) params.set("sub_event_id", filter.subEventId)
  if (filter.search) params.set("search", filter.search)
  const qs = params.toString()
  return qs ? `?${qs}` : ""
}

export async function listPeople(eventId: string, filter: PeopleFilter = {}) {
  return apiRequest(
    { url: `/events/${eventId}/people${toQueryString(filter)}`, method: "GET" },
    peopleListResponseSchema
  )
}

export async function getPerson(eventId: string, id: string) {
  return apiRequest(
    { url: `/events/${eventId}/people/${id}`, method: "GET" },
    personResponseSchema
  )
}

export async function createPerson(eventId: string, input: CreatePersonInput) {
  const body = createPersonSchema.parse(input)
  return apiRequest(
    { url: `/events/${eventId}/people`, method: "POST", data: body },
    personResponseSchema
  )
}

export async function updatePerson(
  eventId: string,
  id: string,
  input: UpdatePersonInput
) {
  const body = updatePersonSchema.parse(input)
  return apiRequest(
    { url: `/events/${eventId}/people/${id}`, method: "PATCH", data: body },
    personResponseSchema
  )
}

export async function deletePerson(eventId: string, id: string) {
  await apiRequest(
    { url: `/events/${eventId}/people/${id}`, method: "DELETE" },
    z.unknown()
  )
}

export async function importPeopleCsv(eventId: string, formData: FormData) {
  return apiRequest(
    { url: `/events/${eventId}/people/import`, method: "POST", data: formData },
    importSummarySchema
  )
}