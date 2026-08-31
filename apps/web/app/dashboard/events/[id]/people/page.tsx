"use client";

import Link from "next/link";
import { notFound, useParams, useRouter, useSearchParams } from "next/navigation";

import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { useEventDetailQuery } from "@/query-hooks/events.api";
import { usePeopleListQuery } from "@/query-hooks/people.api";
import { useSubEventsListQuery } from "@/query-hooks/subevents.api";
import { ApiError } from "@/react-query/client";

export default function PeoplePage() {
    const { id } = useParams<{ id: string }>();
    const router = useRouter();
    const searchParams = useSearchParams();
    const subEventId = searchParams.get("sub_event_id") ?? undefined;
    const search = searchParams.get("search") ?? undefined;

    const { data: event, error } = useEventDetailQuery(id);
    const { data: subEvents = [] } = useSubEventsListQuery(id);
    const { data: people = [] } = usePeopleListQuery(id, { subEventId, search });

    if (error instanceof ApiError && error.status === 404) notFound();

    if (!event) {
        return (
            <div className="flex items-center justify-center py-16 text-sm text-muted-foreground">
                Loading…
            </div>
        );
    }

    const subEventNames = new Map(subEvents.map((se) => [se.id, se.name]));

    const exportParams = new URLSearchParams();
    if (subEventId) exportParams.set("sub_event_id", subEventId);
    if (search) exportParams.set("search", search);

    function applyFilters(e: React.FormEvent<HTMLFormElement>) {
        e.preventDefault();
        const fd = new FormData(e.currentTarget);
        const params = new URLSearchParams();
        const nextSubEvent = fd.get("sub_event_id")?.toString() || "";
        const nextSearch = fd.get("search")?.toString() || "";
        if (nextSubEvent) params.set("sub_event_id", nextSubEvent);
        if (nextSearch) params.set("search", nextSearch);
        const qs = params.toString();
        router.push(`/dashboard/events/${id}/people${qs ? `?${qs}` : ""}`);
    }

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

            <form onSubmit={applyFilters} className="flex items-center gap-2">
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
                <div className="overflow-x-auto rounded-lg border">
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
    );
}