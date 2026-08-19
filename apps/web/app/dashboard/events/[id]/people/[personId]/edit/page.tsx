"use client"

import { notFound, useParams, useRouter } from "next/navigation"
import { useQuery } from "@tanstack/react-query"

import { ApiError } from "@/react-query/client"
import { getEvent } from "@/lib/client-api/events"
import { deletePerson, getPerson, updatePerson } from "@/lib/client-api/people"
import { listSubEvents } from "@/lib/client-api/subevents"
import { queryKeys } from "@/react-query/query-keys"

import { PersonForm } from "../../person-form"
import { CardActions } from "./card-actions"
import { DeletePersonButton } from "./delete-person-button"

export default function EditPersonPage() {
  const { id, personId } = useParams<{ id: string; personId: string }>()
  const router = useRouter()

  const { data: person, error } = useQuery({
    queryKey: queryKeys.person(id, personId),
    queryFn: () => getPerson(id, personId),
    retry: false,
  })

  const { data: event } = useQuery({
    queryKey: queryKeys.event(id),
    queryFn: () => getEvent(id),
  })

  const { data: subEvents = [] } = useQuery({
    queryKey: queryKeys.subEvents(id),
    queryFn: () => listSubEvents(id),
  })

  if (error instanceof ApiError && error.status === 404) notFound()

  if (!person || !event) {
    return (
      <div className="flex items-center justify-center py-16 text-sm text-muted-foreground">
        Loading…
      </div>
    )
  }

  async function updateAction(
    values: Parameters<typeof updatePerson>[2]
  ): Promise<{ error: string } | undefined> {
    try {
      await updatePerson(id, personId, values)
    } catch (err: any) {
      return { error: err.message ?? "Something went wrong." }
    }
    router.push(`/dashboard/events/${id}/people`)
  }

  async function deleteAction(): Promise<{ error: string } | undefined> {
    try {
      await deletePerson(id, personId)
    } catch (err: any) {
      return { error: err.message ?? "Something went wrong." }
    }
    router.push(`/dashboard/events/${id}/people`)
  }

  return (
    <div className="flex flex-col gap-6">
      <div className="flex items-center justify-between">
        <h1 className="font-heading text-xl font-medium">Edit {person.name}</h1>
        <DeletePersonButton action={deleteAction} />
      </div>
      <CardActions
        eventId={id}
        personId={personId}
        downloadHref={`/dashboard/events/${id}/people/${personId}/card`}
        eventPublished={event.status === "published"}
        cardSentAt={person.card_sent_at}
      />
      <PersonForm
        subEvents={subEvents}
        showIdentity={false}
        defaultValues={{
          email: person.email,
          mobile: person.mobile,
          name: person.name,
          image_url: person.image_url ?? "",
          age: person.age?.toString() ?? "",
          gender: person.gender ?? "",
          sub_event_ids: person.sub_event_ids,
        }}
        action={updateAction}
        submitLabel="Save changes"
      />
    </div>
  )
}