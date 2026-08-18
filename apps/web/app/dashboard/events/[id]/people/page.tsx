import Link from "next/link"
import { notFound } from "next/navigation"

import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { ApiError } from "@/lib/api/client"
import { getEvent } from "@/lib/api/events"
import { listPeople } from "@/lib/api/people"
import { listSubEvents } from "@/lib/api/subevents"

export default async function PeoplePage({
  params,
  searchParams,
}: {
  params: Promise<{ id: string }>
  searchParams: Promise<{ sub_event_id?: string; search?: string }>
}) {
  const { id } = await params
  const { sub_event_id: subEventId, search } = await searchParams

  let event
  try {
    event = await getEvent(id)
  } catch (err) {
    if (err instanceof ApiError && err.status === 404) notFound()
    throw err
  }
  const [subEvents, people] = await Promise.all([
    listSubEvents(id),
    listPeople(id, { subEventId, search }),
  ])
  const subEventNames = new Map(subEvents.map((se) => [se.id, se.name]))

  const exportParams = new URLSearchParams()
  if (subEventId) exportParams.set("sub_event_id", subEventId)
  if (search) exportParams.set("search", search)

  return (
    <div className="flex flex-col gap-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="font-heading text-xl font-medium">
            People — {event.name}
          </h1>
          <p className="text-sm text-muted-foreground">
            {people.length} {people.length === 1 ? "person" : "people"}
          </p>
        </div>
        <div className="flex items-center gap-2">
          <Button
            variant="outline"
            render={<Link href={`/dashboard/events/${id}/forms`} />}
          >
            Public forms
          </Button>
          <Button
            variant="outline"
            render={<Link href={`/dashboard/events/${id}/people/import`} />}
          >
            Import CSV
          </Button>
          <Button
            variant="outline"
            render={
              <a
                href={`/dashboard/events/${id}/people/export?${exportParams.toString()}`}
              />
            }
          >
            Export CSV
          </Button>
          <Button render={<Link href={`/dashboard/events/${id}/people/new`} />}>
            Add person
          </Button>
        </div>
      </div>

      <form method="get" className="flex items-center gap-2">
        <Input
          name="search"
          placeholder="Search name, email, mobile"
          defaultValue={search}
          className="max-w-xs"
        />
        <select
          name="sub_event_id"
          defaultValue={subEventId ?? ""}
          className="h-8 rounded-lg border border-input bg-transparent px-2.5 text-sm outline-none focus-visible:border-ring focus-visible:ring-3 focus-visible:ring-ring/50 dark:bg-input/30"
        >
          <option value="">All sub-events</option>
          {subEvents.map((se) => (
            <option key={se.id} value={se.id}>
              {se.name}
            </option>
          ))}
        </select>
        <Button type="submit" variant="outline" size="sm">
          Filter
        </Button>
      </form>

      {people.length === 0 ? (
        <p className="text-sm text-muted-foreground">No people yet.</p>
      ) : (
        <div className="overflow-x-auto rounded-xl border">
          <table className="w-full text-sm">
            <thead className="border-b bg-muted/50 text-left text-muted-foreground">
              <tr>
                <th className="p-2 font-medium">Name</th>
                <th className="p-2 font-medium">Email</th>
                <th className="p-2 font-medium">Mobile</th>
                <th className="p-2 font-medium">Sub-events</th>
                <th className="p-2 font-medium">Joined</th>
                <th className="p-2" />
              </tr>
            </thead>
            <tbody className="divide-y">
              {people.map((person) => (
                <tr key={person.id}>
                  <td className="p-2">{person.name}</td>
                  <td className="p-2">{person.email}</td>
                  <td className="p-2">{person.mobile}</td>
                  <td className="p-2 text-muted-foreground">
                    {person.sub_event_ids.length === 0
                      ? "—"
                      : person.sub_event_ids
                          .map((sid) => subEventNames.get(sid) ?? sid)
                          .join(", ")}
                  </td>
                  <td className="p-2 text-muted-foreground">
                    {person.joined_at ? "Joined" : "Pending publish"}
                  </td>
                  <td className="p-2 text-right">
                    <Link
                      href={`/dashboard/events/${id}/people/${person.id}/edit`}
                      className="text-sm underline underline-offset-4"
                    >
                      Edit
                    </Link>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </div>
  )
}
