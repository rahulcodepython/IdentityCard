import { CreateEventForm } from "./create-event-form"

export default function NewEventPage() {
  return (
    <div className="flex flex-col gap-6">
      <h1 className="font-heading text-xl font-medium">New event</h1>
      <CreateEventForm />
    </div>
  )
}
