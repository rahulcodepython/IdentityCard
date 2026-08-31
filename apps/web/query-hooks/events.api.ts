"use client";

import { apiRequest } from "@/react-query/client";
import {
    removeFromArray,
    useArrayMutation,
    useObjectMutation,
    useSimpleMutation,
    useWithExecute,
} from "@/react-query/mutation";
import { useAppQuery } from "@/react-query/query";
import { queryKeys } from "@/react-query/query-keys";
import {
    type CreateEventInput,
    type DayImportSummary,
    type EventDetail,
    type EventSummary,
    type UpdateEventInput,
    dayImportSummarySchema,
    eventResponseSchema,
    eventsListResponseSchema,
} from "@/schema/events.types";
import { z } from "zod";

export function useEventsListQuery(enabled = true) {
    return useAppQuery<EventSummary[]>(
        queryKeys.events(),
        () =>
            apiRequest(
                { url: "/events", method: "GET" },
                eventsListResponseSchema
            ),
        { enabled }
    );
}

export function useEventDetailQuery(eventId: string, enabled = true) {
    return useAppQuery<EventDetail>(
        queryKeys.event(eventId),
        () =>
            apiRequest(
                { url: `/events/${eventId}`, method: "GET" },
                eventResponseSchema
            ),
        { enabled: enabled && !!eventId }
    );
}

export function useCreateEventMutation() {
    return useWithExecute(
        useArrayMutation<EventSummary, EventDetail, CreateEventInput>({
            mutationFn: (data) =>
                apiRequest(
                    {
                        url: "/events",
                        method: "POST",
                        data,
                    },
                    eventResponseSchema
                ),
            queryKey: queryKeys.events(),
            updater: (created) => (old) => [
                ...(old ?? []),
                {
                    id: created.id,
                    name: created.name,
                    event_type: created.event_type,
                    status: created.status,
                    start_date: created.start_date,
                    end_date: created.end_date,
                    venue: created.venue,
                    organizer_name: created.organizer_name,
                    has_image: created.has_image,
                    has_organizer_signature: created.has_organizer_signature,
                    published_at: created.published_at,
                },
            ],
            invalidateKeys: [queryKeys.orgSubscriptions()],
            showToast: { success: "Event created successfully" },
        })
    );
}

export function useUpdateEventMutation(eventId: string) {
    return useWithExecute(
        useObjectMutation<EventDetail, UpdateEventInput>({
            mutationFn: (data) =>
                apiRequest(
                    {
                        url: `/events/${eventId}`,
                        method: "PATCH",
                        data,
                    },
                    eventResponseSchema
                ),
            queryKey: queryKeys.event(eventId),
            updater: (updated) => () => updated,
            invalidateKeys: [queryKeys.events()],
            showToast: { success: "Event updated successfully" },
        })
    );
}

export function usePublishEventMutation(eventId: string) {
    return useWithExecute(
        useObjectMutation<EventDetail, void>({
            mutationFn: () =>
                apiRequest(
                    {
                        url: `/events/${eventId}/publish`,
                        method: "POST",
                    },
                    eventResponseSchema
                ),
            queryKey: queryKeys.event(eventId),
            updater: (updated) => () => updated,
            invalidateKeys: [queryKeys.events()],
            showToast: { success: "Event published successfully" },
        })
    );
}

export function useDeleteEventMutation(eventId: string) {
    return useWithExecute(
        useArrayMutation<EventSummary, void, void>({
            mutationFn: () =>
                apiRequest(
                    {
                        url: `/events/${eventId}`,
                        method: "DELETE",
                    },
                    z.any()
                ),
            queryKey: queryKeys.events(),
            updater: () => removeFromArray(eventId),
            invalidateKeys: [queryKeys.event(eventId), queryKeys.orgSubscriptions()],
            showToast: { success: "Draft event deleted" },
        })
    );
}

export function useImportDaysMutation(eventId: string) {
    return useWithExecute(
        useSimpleMutation<DayImportSummary, FormData>({
            mutationFn: (formData) =>
                apiRequest(
                    {
                        url: `/events/${eventId}/days/import`,
                        method: "POST",
                        data: formData,
                        headers: { "Content-Type": "multipart/form-data" },
                    },
                    dayImportSummarySchema
                ),
            invalidateKeys: [queryKeys.event(eventId), queryKeys.events()],
            showToast: { success: "Schedule days imported" },
        })
    );
}

export function useUploadEventImageMutation(eventId: string) {
    return useWithExecute(
        useSimpleMutation<void, FormData>({
            mutationFn: (formData) =>
                apiRequest(
                    {
                        url: `/events/${eventId}/image`,
                        method: "POST",
                        data: formData,
                        headers: { "Content-Type": "multipart/form-data" },
                    },
                    z.any()
                ),
            invalidateKeys: [queryKeys.event(eventId), queryKeys.events()],
            showToast: { success: "Event banner uploaded" },
        })
    );
}

export function useUploadOrganizerSignatureMutation(eventId: string) {
    return useWithExecute(
        useSimpleMutation<void, FormData>({
            mutationFn: (formData) =>
                apiRequest(
                    {
                        url: `/events/${eventId}/signature`,
                        method: "POST",
                        data: formData,
                        headers: { "Content-Type": "multipart/form-data" },
                    },
                    z.any()
                ),
            invalidateKeys: [queryKeys.event(eventId), queryKeys.events()],
            showToast: { success: "Signature uploaded" },
        })
    );
}

