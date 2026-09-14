"use client";

import * as React from "react";

import { CreateEventDialog } from "@/components/events/create-event-dialog";
import { eventsColumns } from "@/components/events/events-columns";
import { InfiniteDataTable } from "@/components/generic";
import { useBreadcrumbs } from "@/hooks/use-breadcrumbs";
import { useEventsInfiniteQuery } from "@/query-hooks/events.api";

export default function EventsPage() {
    const [search, setSearch] = React.useState("");

    useBreadcrumbs([
        {
            title: "Dashboard",
            url: "/dashboard",
        },
        {
            title: "Events",
        },
    ]);

    const {
        data,
        isLoading,
        isError,
        error,
        fetchNextPage,
        hasNextPage,
        isFetchingNextPage,
    } = useEventsInfiniteQuery({ search });

    const flatEvents = React.useMemo(() => {
        return data?.pages.flatMap((page) => page.data) ?? [];
    }, [data]);

    const totalCount = data?.pages[0]?.total ?? flatEvents.length;

    return (
        <InfiniteDataTable
            columns={eventsColumns}
            data={flatEvents}
            totalCount={totalCount}
            isLoading={isLoading}
            isError={isError}
            error={error}
            hasNextPage={hasNextPage}
            isFetchingNextPage={isFetchingNextPage}
            fetchNextPage={fetchNextPage}
            searchPlaceholder="Search events..."
            onSearchChange={setSearch}
            toolbarActions={<CreateEventDialog />}
            itemLabel="events"
        />
    );
}
