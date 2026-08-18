import { cookies } from "next/headers"
import { type NextRequest } from "next/server"

// The one Route Handler in this app: a CSV download can't go through the
// JSON-envelope apiFetch client, and (being a browser-navigated link, not
// a fetch call) can't attach the httpOnly auth cookie itself either — this
// proxies the request server-side, attaching the cookie and streaming the
// Go API's response straight through.
const API_BASE_URL = process.env.API_BASE_URL ?? "http://localhost:8080"

export async function GET(
  request: NextRequest,
  { params }: { params: Promise<{ id: string }> }
) {
  const { id } = await params
  const cookieStore = await cookies()

  const res = await fetch(
    `${API_BASE_URL}/events/${id}/people/export${request.nextUrl.search}`,
    {
      headers: { Cookie: cookieStore.toString() },
      cache: "no-store",
    }
  )

  if (!res.ok || !res.body) {
    return new Response("Failed to export people", {
      status: res.status || 500,
    })
  }

  return new Response(res.body, {
    status: 200,
    headers: {
      "Content-Type": "text/csv; charset=utf-8",
      "Content-Disposition":
        res.headers.get("Content-Disposition") ??
        'attachment; filename="people.csv"',
    },
  })
}
