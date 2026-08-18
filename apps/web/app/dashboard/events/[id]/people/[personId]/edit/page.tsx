import { notFound } from "next/navigation"

import { ApiError } from "@/lib/api/client"
import { getEvent } from "@/lib/api/events"
import { getPerson } from "@/lib/api/people"
import { listSubEvents } from "@/lib/api/subevents"

import { PersonForm } from "../../person-form"
import { deletePersonAction, updatePersonAction } from "./actions"
import { CardActions } from "./card-actions"
import { DeletePersonButton } from "./delete-person-button"

export default async function EditPersonPage({
  params,
}: {
  params: Promise<{ id: string; personId: string }>
}) {
  const { id, personId } = await params

  let person
  try {
    person = await getPerson(id, personId)
  } catch (err) {
    if (err instanceof ApiError && err.status === 404) notFound()
    throw err
  }
  const [event, subEvents] = await Promise.all([
    getEvent(id),
    listSubEvents(id),
  ])

  const updateAction = updatePersonAction.bind(null, id, personId)
  const deleteAction = deletePersonAction.bind(null, id, personId)

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
