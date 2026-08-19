"use client"

import Link from "next/link"
import type { ColumnDef } from "@tanstack/react-table"

import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { DataTable } from "@/components/ui/data-table"
import type { EventSummary } from "@/lib/validation/events"

const SCHEDULE_MODE_LABEL: Record<EventSummary["schedule_mode"], string> = {
  flash: "Flash",
  fixed_range: "Fixed range",
  selective: "Selective",
  recurring: "Recurring",
}

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
    accessorKey: "schedule_mode",
    header: "Schedule",
    cell: ({ row }) => (
      <Badge variant="outline">{SCHEDULE_MODE_LABEL[row.original.schedule_mode]}</Badge>
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
    cell: ({ row }) => row.original.end_date ?? "Open-ended",
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
]

export function EventsTable({ data }: { data: EventSummary[] }) {
  return (
    <DataTable
      columns={columns}
      data={data}
      searchPlaceholder="Search events..."
      emptyMessage="No events yet."
    />
  )
}
