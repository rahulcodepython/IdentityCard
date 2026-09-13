"use client";

import { notFound, useParams } from "next/navigation";

import { useSetBreadcrumbs } from "@/hooks/use-breadcrumbs";
import { useEventDetailQuery } from "@/query-hooks/events.api";
import { ApiError } from "@/react-query/client";

export default function EventLayout({ children }: { children: React.ReactNode }) {
    const { id, orgSlug } = useParams<{ id: string; orgSlug: string }>();

    const { data: event, error } = useEventDetailQuery(id);

    if (error instanceof ApiError && error.status === 404) notFound();

    useSetBreadcrumbs([
        { label: "Events", href: `/dashboard/${orgSlug}/events` },
        { label: event?.name ?? "Event", href: `/dashboard/${orgSlug}/events/${id}` },
    ]);

    return children;
}