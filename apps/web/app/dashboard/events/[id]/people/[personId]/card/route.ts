import { headers } from "next/headers"
import { type NextRequest } from "next/server"

import { auth } from "@/lib/auth"
import { API_V1_PREFIX } from "@/lib/constants"

const API_BASE_URL = process.env.API_BASE_URL ?? "http://localhost:8000"

export async function GET(
    _request: NextRequest,
    { params }: { params: Promise<{ id: string; personId: string }> }
) {
    const { id, personId } = await params
    const { token } = await auth.api.getToken({ headers: await headers() })

    const res = await fetch(
        `${API_BASE_URL}${API_V1_PREFIX}/events/${id}/people/${personId}/card`,
        {
            headers: { Authorization: `Bearer ${token}` },
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
