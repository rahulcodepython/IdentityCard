"use client"

import { useQuery } from "@tanstack/react-query"
import { useParams, notFound } from "next/navigation"

import { BreadcrumbSetter } from "@/components/breadcrumb-setter"
import { ApiError } from "@/react-query/client"
import { getEvent } from "@/lib/client-api/events"
import { queryKeys } from "@/react-query/query-keys"

export default function EventLayout({ children }: { children: React.ReactNode }) {
  const { id } = useParams<{ id: string }>()

  const { data: event, error } = useQuery({
    queryKey: queryKeys.event(id),
    queryFn: () => getEvent(id),
    retry: false,
  })

  if (error instanceof ApiError && error.status === 404) notFound()

  return (
    <>
      {event && <BreadcrumbSetter id={id} label={event.name} />}
      {children}
    </>
  )
}