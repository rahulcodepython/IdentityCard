import Link from "next/link"

import { Button } from "@/components/ui/button"
import {
  Card,
  CardAction,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card"
import { listEvents } from "@/lib/api/events"

export default async function EventsPage() {
  const events = await listEvents()

  return (
    <div className="flex flex-col gap-6">
      <div className="flex items-center justify-between">
        <h1 className="font-heading text-xl font-medium">Events</h1>
        <Button render={<Link href="/dashboard/events/new" />}>
          New event
        </Button>
      </div>

      {events.length === 0 ? (
        <p className="text-sm text-muted-foreground">No events yet.</p>
      ) : (
        <div className="flex flex-col gap-3">
          {events.map((event) => (
            <Card key={event.id}>
              <CardHeader>
                <CardTitle>
                  <Link
                    href={`/dashboard/events/${event.id}`}
                    className="hover:underline"
                  >
                    {event.name}
                  </Link>
                </CardTitle>
                <CardDescription>
                  {event.kind === "flash" ? "Flash" : "Established"} ·{" "}
                  {event.start_date}
                  {event.end_date !== event.start_date
                    ? ` – ${event.end_date}`
                    : ""}{" "}
                  · {event.status === "published" ? "Published" : "Draft"}
                </CardDescription>
                <CardAction>
                  <Button
                    variant="outline"
                    render={<Link href={`/dashboard/events/${event.id}`} />}
                  >
                    View
                  </Button>
                </CardAction>
              </CardHeader>
            </Card>
          ))}
        </div>
      )}
    </div>
  )
}
