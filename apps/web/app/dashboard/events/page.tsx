"use client";

import * as React from "react";
import { Loader2, SearchIcon } from "lucide-react";

import { CreateEventDialog } from "@/components/events/create-event-dialog";
import { eventsColumns } from "@/components/events/events-columns";
import { DataTable } from "@/components/ui/data-table";
import { Input } from "@/components/ui/input";
import { useEventsInfiniteQuery } from "@/query-hooks/events.api";

export default function EventsPage() {
    const [searchInput, setSearchInput] = React.useState("");
    const [search, setSearch] = React.useState("");
    const observerTarget = React.useRef<HTMLDivElement>(null);

    // Debounce search query input by 300ms
    React.useEffect(() => {
        const handler = setTimeout(() => {
            setSearch(searchInput.trim());
        }, 300);
        return () => clearTimeout(handler);
    }, [searchInput]);

    const {
        data,
        isLoading,
        isError,
        error,
        fetchNextPage,
        hasNextPage,
        isFetchingNextPage,
    } = useEventsInfiniteQuery({ search });

    // Automatically trigger fetchNextPage when user scrolls near the bottom
    React.useEffect(() => {
        const target = observerTarget.current;
        if (!target) return;

        const observer = new IntersectionObserver(
            (entries) => {
                const [entry] = entries;
                if (entry?.isIntersecting && hasNextPage && !isFetchingNextPage) {
                    fetchNextPage();
                }
            },
            {
                root: null,
                rootMargin: "200px",
                threshold: 0,
            }
        );

        observer.observe(target);
        return () => observer.disconnect();
    }, [fetchNextPage, hasNextPage, isFetchingNextPage]);

    const flatEvents = React.useMemo(() => {
        return data?.pages.flatMap((page) => page.data) ?? [];
    }, [data]);

    const totalCount = data?.pages[0]?.total ?? flatEvents.length;

    return (
        <div className="flex flex-1 flex-col gap-4">
            {/* Top Toolbar: Search Bar on left & Create Button on top-right */}
            <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
                <div className="relative w-full max-w-sm">
                    <SearchIcon className="absolute left-2.5 top-1/2 size-4 -translate-y-1/2 text-muted-foreground" />
                    <Input
                        type="text"
                        placeholder="Search events..."
                        value={searchInput}
                        onChange={(e) => setSearchInput(e.target.value)}
                        className="pl-8"
                    />
                </div>
                <div className="flex items-center justify-end">
                    <CreateEventDialog />
                </div>
            </div>

            {/* Error banner */}
            {isError ? (
                <div className="rounded-md border border-destructive/20 bg-destructive/10 p-3 text-xs text-destructive">
                    Failed to load events: {error?.message || "Unknown error"}
                </div>
            ) : null}

            {/* Table structure */}
            <DataTable
                columns={eventsColumns}
                data={flatEvents}
                isLoading={isLoading}
            />

            {/* Counter info */}
            <div className="flex items-center justify-between text-xs text-muted-foreground">
                <span>
                    Showing {flatEvents.length} of {totalCount} events
                </span>
            </div>

            {/* Infinite Scroll Sentinel & Loader */}
            <div
                ref={observerTarget}
                className="flex items-center justify-center py-4 text-xs text-muted-foreground"
            >
                {isFetchingNextPage ? (
                    <div className="flex items-center gap-2">
                        <Loader2 className="size-4 animate-spin text-primary" />
                        <span>Loading more events...</span>
                    </div>
                ) : hasNextPage ? (
                    <span className="text-muted-foreground/60">Scroll down to load more</span>
                ) : flatEvents.length > 0 ? (
                    <span>All events loaded</span>
                ) : null}
            </div>
        </div>
    );
}
