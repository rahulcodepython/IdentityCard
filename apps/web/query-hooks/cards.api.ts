"use client";

import { apiRequest } from "@/react-query/client";
import {
    useSimpleMutation,
    useWithExecute,
} from "@/react-query/mutation";
import { queryKeys } from "@/react-query/query-keys";
import {
    type ResendResponse,
    resendResponseSchema,
} from "@/schema/cards.types";

export function useResendCardMutation(eventId: string) {
    return useWithExecute(
        useSimpleMutation<ResendResponse, string>({
            mutationFn: (personId) =>
                apiRequest(
                    {
                        url: `/events/${eventId}/people/${personId}/cards/resend`,
                        method: "POST",
                    },
                    resendResponseSchema
                ),
            invalidateKeys: [queryKeys.person(eventId, ""), queryKeys.people(eventId)],
            showToast: { success: "ID card email resent" },
        })
    );
}

