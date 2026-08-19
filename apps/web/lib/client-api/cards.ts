import { apiRequest } from "@/react-query/client"
import { resendResponseSchema } from "@/schema/cards.types"

export async function resendCard(eventId: string, personId: string) {
  return apiRequest(
    {
      url: `/events/${eventId}/people/${personId}/card/resend`,
      method: "POST",
    },
    resendResponseSchema
  )
}