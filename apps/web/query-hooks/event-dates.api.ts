"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";

import { apiRequest } from "@/react-query/client";
import { queryKeys } from "@/react-query/query-keys";
import {
    BulkDeleteEventDatesSchema,
    BulkUpsertEventDatesSchema,
    EventDatesListResponseSchema,
    type BulkDeleteEventDatesInput,
    type BulkUpsertEventDatesInput,
    type EventDate,
    type EventDateItemInput,
} from "@/schema/event-dates.types";

// Fetch all event dates for a given month in a single request
export async function getEventDates(
    eventId: string,
    month?: string
): Promise<EventDate[]> {
    try {
        const data = await apiRequest<EventDate[]>(
            {
                url: `/events/${eventId}/dates`,
                method: "GET",
                params: month ? { month } : undefined,
            },
            EventDatesListResponseSchema
        );
        return data ?? [];
    } catch {
        return [];
    }
}

// React-query hook for event dates within a month
export function useEventDatesQuery(eventId: string, month: string) {
    return useQuery<EventDate[], Error>({
        queryKey: queryKeys.eventDates.byMonth(eventId, month),
        queryFn: async () => {
            const data = await getEventDates(eventId, month);
            return data ?? [];
        },
        enabled: Boolean(eventId),
    });
}

// React-query hook for all event dates
export function useAllEventDatesQuery(eventId: string) {
    return useQuery<EventDate[], Error>({
        queryKey: queryKeys.eventDates.byEvent(eventId),
        queryFn: async () => {
            const data = await getEventDates(eventId);
            return data ?? [];
        },
        enabled: Boolean(eventId),
    });
}

// Bulk save (upsert) event dates mutation
export function useBulkSaveEventDatesMutation(eventId: string) {
    const queryClient = useQueryClient();

    return useMutation<EventDate[], Error, BulkUpsertEventDatesInput>({
        mutationFn: async (payload: BulkUpsertEventDatesInput) => {
            const validated = BulkUpsertEventDatesSchema.parse(payload);
            return apiRequest<EventDate[]>(
                {
                    url: `/events/${eventId}/dates/bulk`,
                    method: "POST",
                    data: validated,
                },
                EventDatesListResponseSchema
            );
        },
        onSuccess: () => {
            queryClient.invalidateQueries({
                queryKey: queryKeys.eventDates.byEvent(eventId),
            });
            toast.success("Event dates saved successfully");
        },
        onError: (error) => {
            toast.error(error.message || "Failed to save event dates");
        },
    });
}

// Bulk delete event dates mutation
export function useBulkDeleteEventDatesMutation(eventId: string) {
    const queryClient = useQueryClient();

    return useMutation<unknown, Error, BulkDeleteEventDatesInput>({
        mutationFn: async (payload: BulkDeleteEventDatesInput) => {
            const validated = BulkDeleteEventDatesSchema.parse(payload);
            return apiRequest(
                {
                    url: `/events/${eventId}/dates/bulk`,
                    method: "DELETE",
                    data: validated,
                }
            );
        },
        onSuccess: () => {
            queryClient.invalidateQueries({
                queryKey: queryKeys.eventDates.byEvent(eventId),
            });
            toast.success("Event dates deleted successfully");
        },
        onError: (error) => {
            toast.error(error.message || "Failed to delete event dates");
        },
    });
}

// Overwride all event dates mutation (used for file uploads)
export function useOverwrideEventDatesMutation(eventId: string) {
    const queryClient = useQueryClient();

    return useMutation<EventDate[], Error, BulkUpsertEventDatesInput>({
        mutationFn: async (payload: BulkUpsertEventDatesInput) => {
            const validated = BulkUpsertEventDatesSchema.parse(payload);
            return apiRequest<EventDate[]>(
                {
                    url: `/events/${eventId}/dates/overwride`,
                    method: "POST",
                    data: validated,
                },
                EventDatesListResponseSchema
            );
        },
        onSuccess: () => {
            queryClient.invalidateQueries({
                queryKey: queryKeys.eventDates.byEvent(eventId),
            });
            toast.success("All event dates overwridden successfully");
        },
        onError: (error) => {
            toast.error(error.message || "Failed to overwride event dates");
        },
    });
}

// Alias for backwards compatibility
export const useReplaceAllEventDatesMutation = useOverwrideEventDatesMutation;

export interface BulkUpdateEventDatesInput {
    upsertDates?: EventDateItemInput[];
    deleteDates?: string[];
}

// Unified bulk update (atomic delete + upsert without race conditions or duplicate toasts)
export function useBulkUpdateEventDatesMutation(eventId: string) {
    const queryClient = useQueryClient();

    return useMutation<EventDate[], Error, BulkUpdateEventDatesInput>({
        mutationFn: async (payload: BulkUpdateEventDatesInput) => {
            if (payload.deleteDates && payload.deleteDates.length > 0) {
                const validatedDelete = BulkDeleteEventDatesSchema.parse({
                    dates: payload.deleteDates,
                });
                await apiRequest({
                    url: `/events/${eventId}/dates/bulk`,
                    method: "DELETE",
                    data: validatedDelete,
                });
            }

            let result: EventDate[] = [];
            if (payload.upsertDates && payload.upsertDates.length > 0) {
                const validatedUpsert = BulkUpsertEventDatesSchema.parse({
                    dates: payload.upsertDates,
                });
                result = await apiRequest<EventDate[]>(
                    {
                        url: `/events/${eventId}/dates/bulk`,
                        method: "POST",
                        data: validatedUpsert,
                    },
                    EventDatesListResponseSchema
                );
            }
            return result;
        },
        onSuccess: () => {
            queryClient.invalidateQueries({
                queryKey: queryKeys.eventDates.byEvent(eventId),
            });
            toast.success("Event dates saved successfully");
        },
        onError: (error) => {
            toast.error(error.message || "Failed to update event dates");
        },
    });
}


