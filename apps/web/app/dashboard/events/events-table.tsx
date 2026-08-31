"use client";

import Link from "next/link";
import type { ColumnDef } from "@tanstack/react-table";

import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { DataTable } from "@/components/data-table";
import type { EventSummary } from "@/schema/events.types";

const EVENT_TYPE_LABEL: Record<EventSummary["event_type"], string> = {
    flash: "Flash",
    standard: "Standard",
    grouped: "Grouped",
};

const columns: ColumnDef<EventSummary>[] = [
    {
        accessorKey: "name",
        header: "Name",
        enableHiding: false,
        cell: ({ row }) => (
            <Link
                href={`/dashboard/events/${row.original.id}`}
                className="font-medium hover:underline"
            >
                {row.original.name}
            </Link>
        ),
    },
    {
        accessorKey: "event_type",
        header: "Type",
        cell: ({ row }) => (
            <Badge variant="outline" className="capitalize">
                {EVENT_TYPE_LABEL[row.original.event_type] || row.original.event_type}
            </Badge>
        ),
    },
    {
        accessorKey: "status",
        header: "Status",
        cell: ({ row }) => (
            <Badge variant={row.original.status === "published" ? "default" : "secondary"}>
                {row.original.status === "published" ? "Published" : "Draft"}
            </Badge>
        ),
    },
    {
        accessorKey: "start_date",
        header: "Start date",
    },
    {
        accessorKey: "end_date",
        header: "End date",
        cell: ({ row }) => row.original.end_date || "—",
    },
    {
        accessorKey: "venue",
        header: "Venue",
        cell: ({ row }) => row.original.venue || "—",
    },
    {
        id: "actions",
        header: "",
        enableHiding: false,
        enableSorting: false,
        cell: ({ row }) => (
            <Button
                variant="outline"
                size="sm"
                render={<Link href={`/dashboard/events/${row.original.id}`} />}
            >
                View
            </Button>
        ),
    },
];

export function EventsTable({ data }: { data: EventSummary[] }) {
    return (
        <DataTable
            columns={columns}
            data={data}
            searchPlaceholder="Search events..."
            emptyMessage="No events yet."
        />
    );
}
