import "server-only"

import { apiFetch } from "@/lib/api/client"
import { resendResponseSchema } from "@/lib/validation/cards"

export async function resendCard(eventId: string, personId: string) {
  const data = await apiFetch(
    `/events/${eventId}/people/${personId}/card/resend`,
    {
      method: "POST",
    }
  )
  return resendResponseSchema.parse(data)
}
