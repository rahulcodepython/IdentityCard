import { notFound } from "next/navigation"

import { BreadcrumbSetter } from "@/components/breadcrumb-setter"
import { ApiError } from "@/lib/api/client"
import { getEvent } from "@/lib/api/events"

export default async function EventLayout({
  children,
  params,
}: {
  children: React.ReactNode
  params: Promise<{ id: string }>
}) {
  const { id } = await params

  const event = await getEvent(id).catch((err: unknown) => {
    if (err instanceof ApiError && err.status === 404) notFound()
    throw err
  })

  return (
    <>
      <BreadcrumbSetter id={id} label={event.name} />
      {children}
    </>
  )
}
