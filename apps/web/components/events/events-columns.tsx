"use client";

import { type ColumnDef } from "@tanstack/react-table";

import { Badge } from "@/components/ui/badge";
import type { Event } from "@/schema/events.types";
import Link from "next/link";

export const eventsColumns: ColumnDef<Event>[] = [
    {
        accessorKey: "name",
        header: "Event Name",
        cell: ({ row }) => (
            <Link className="flex flex-col" href={"/dashboard/events/" + row.original.id}>
                <span className="font-medium text-foreground">{row.original.name}</span>
            </Link>
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
