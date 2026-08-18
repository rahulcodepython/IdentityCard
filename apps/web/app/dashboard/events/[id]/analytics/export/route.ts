import { cookies } from "next/headers"
import { type NextRequest } from "next/server"

// Same pattern as the people CSV export: a browser-navigated download
// link can't attach the httpOnly auth cookie itself, so this proxies the
// request server-side and streams the Go API's CSV response through.
const API_BASE_URL = process.env.API_BASE_URL ?? "http://localhost:8080"

export async function GET(
  request: NextRequest,
  { params }: { params: Promise<{ id: string }> }
) {
  const { id } = await params
  const cookieStore = await cookies()

  const res = await fetch(
    `${API_BASE_URL}/events/${id}/attendance/export${request.nextUrl.search}`,
    {
      headers: { Cookie: cookieStore.toString() },
      cache: "no-store",
    }
  )

  if (!res.ok || !res.body) {
    return new Response("Failed to export attendance", {
      status: res.status || 500,
    })
  }

  return new Response(res.body, {
    status: 200,
    headers: {
      "Content-Type": "text/csv; charset=utf-8",
      "Content-Disposition":
        res.headers.get("Content-Disposition") ??
        'attachment; filename="attendance.csv"',
    },
  })
}
