"use client";

import { type ColumnDef } from "@tanstack/react-table";

import { Badge } from "@/components/ui/badge";
import type { Event } from "@/schema/events.types";

export const eventsColumns: ColumnDef<Event>[] = [
    {
        accessorKey: "name",
        header: "Event Name",
        cell: ({ row }) => (
            <div className="flex flex-col">
                <span className="font-medium text-foreground">{row.original.name}</span>
                {row.original.venue ? (
                    <span className="text-xs text-muted-foreground">{row.original.venue}</span>
                ) : null}
            </div>
        ),
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
        accessorKey: "status",
        header: "Status",
        cell: ({ row }) => {
            const isPublished = row.original.status === "published";
            return (
                <Badge variant={isPublished ? "default" : "secondary"} className="capitalize">
                    {row.original.status}
                </Badge>
            );
        },
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
