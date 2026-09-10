"use client";

import { notFound, useParams } from "next/navigation";

import { BreadcrumbSetter } from "@/components/navigation/breadcrumb-setter";
import { useEventDetailQuery } from "@/query-hooks/events.api";
import { ApiError } from "@/react-query/client";

export default function EventLayout({ children }: { children: React.ReactNode }) {
    const { id } = useParams<{ id: string }>();

    const { data: event, error } = useEventDetailQuery(id);

    if (error instanceof ApiError && error.status === 404) notFound();

    return (
        <>
            {event && <BreadcrumbSetter id={id} label={event.name} />}
            {children}
        </>
    );
}