import { cookies } from "next/headers"
import { type NextRequest } from "next/server"

// Same pattern as the CSV export and org logo proxies: a browser-navigated
// download link can't carry the httpOnly auth cookie itself, and the
// response is a PDF, not the JSON apiFetch expects.
const API_BASE_URL = process.env.API_BASE_URL ?? "http://localhost:8080"

export async function GET(
  _request: NextRequest,
  { params }: { params: Promise<{ id: string; personId: string }> }
) {
  const { id, personId } = await params
  const cookieStore = await cookies()

  const res = await fetch(
    `${API_BASE_URL}/events/${id}/people/${personId}/card`,
    {
      headers: { Cookie: cookieStore.toString() },
      cache: "no-store",
    }
  )

  if (!res.ok || !res.body) {
    return new Response("Failed to generate card", {
      status: res.status || 500,
    })
  }

  return new Response(res.body, {
    status: 200,
    headers: {
      "Content-Type": "application/pdf",
      "Content-Disposition": 'inline; filename="id-card.pdf"',
    },
  })
}
