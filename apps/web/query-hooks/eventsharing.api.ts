import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";

import { apiRequest } from "@/react-query/client";
import { queryKeys } from "@/react-query/query-keys";
import {
    type EventSharingResponse,
    EventSharingResponseSchema,
    type UpdateEventSharingInput,
    UpdateEventSharingSchema,
} from "@/schema/eventsharing.types";

export function useEventSharingQuery(eventId: string) {
    return useQuery<EventSharingResponse, Error>({
        queryKey: queryKeys.eventSharing.detail(eventId),
        queryFn: async () => {
            return apiRequest<EventSharingResponse>(
                {
                    url: `/events/${eventId}/sharing`,
                    method: "GET",
                },
                EventSharingResponseSchema,
            );
        },
        enabled: Boolean(eventId),
    });
}

export function useUpdateEventSharingMutation() {
    const queryClient = useQueryClient();

    return useMutation<
        EventSharingResponse,
        Error,
        { eventId: string } & UpdateEventSharingInput
    >({
        mutationFn: async ({ eventId, ...data }) => {
            const validated = UpdateEventSharingSchema.parse(data);
            return apiRequest<EventSharingResponse>(
                {
                    url: `/events/${eventId}/sharing`,
                    method: "PATCH",
                    data: validated,
                },
                EventSharingResponseSchema,
            );
        },
        onSuccess: (res, variables) => {
            queryClient.setQueryData(
                queryKeys.eventSharing.detail(variables.eventId),
                res,
            );
            toast.success("Sharing settings saved successfully");
        },
        onError: (error) => {
            toast.error(error.message || "Failed to update sharing settings");
        },
    });
}
