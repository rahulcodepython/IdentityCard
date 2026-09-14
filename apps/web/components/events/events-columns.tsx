"use client";

import { type ColumnDef } from "@tanstack/react-table";
import Link from "next/link";
import type { Event } from "@/schema/events.types";

export const eventsColumns: ColumnDef<Event>[] = [
    {
        accessorKey: "name",
        header: "Event Name",
        cell: ({ row }) => {
            return (
                <Link className="flex flex-col" href={`/dashboard/events/${row.original.id}`}>
                    <span className="font-medium text-foreground">
                        {row.original.name}
                    </span>
                </Link>
            );
        },
    },
    {
        accessorKey: "start_date",
        header: "Start Date",
        cell: ({ row }) => <span>{row.original.start_date}</span>,
    },
    {
        accessorKey: "end_date",
        header: "End Date",
        cell: ({ row }) => <span>{row.original.end_date}</span>,
    },
    {
        accessorKey: "created_at",
        header: "Created",
        cell: ({ row }) => {
            const date = new Date(row.original.created_at);
            return (
                <span className="text-xs text-muted-foreground">
                    {isNaN(date.getTime()) ? "-" : date.toLocaleDateString()}
                </span>
            );
        },
    },
];
