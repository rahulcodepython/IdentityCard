"use client";

import { apiRequest } from "@/react-query/client";
import {
    appendToArray,
    removeFromArray,
    replaceInArray,
    useArrayMutation,
    useWithExecute,
} from "@/react-query/mutation";
import { useAppQuery } from "@/react-query/query";
import { queryKeys } from "@/react-query/query-keys";
import {
    type CreateSubEventInput,
    type SubEvent,
    type UpdateSubEventInput,
    subEventResponseSchema,
    subEventsListResponseSchema,
} from "@/schema/subevents.types";
import {
    MSG_SUBEVENT_CREATED,
    MSG_SUBEVENT_DELETED,
    MSG_SUBEVENT_UPDATED,
} from "@/lib/constants";
import { z } from "zod";

export function useSubEventsListQuery(eventId: string, enabled = true) {
    return useAppQuery<SubEvent[]>(
        queryKeys.subEvents(eventId),
        () =>
            apiRequest(
                { url: `/events/${eventId}/subevents`, method: "GET" },
                subEventsListResponseSchema
            ),
        { enabled: enabled && !!eventId }
    );
}

export function useCreateSubEventMutation(eventId: string) {
    return useWithExecute(
        useArrayMutation<SubEvent, SubEvent, CreateSubEventInput>({
            mutationFn: (data) =>
                apiRequest(
                    {
                        url: `/events/${eventId}/subevents`,
                        method: "POST",
                        data,
                    },
                    subEventResponseSchema
                ),
            queryKey: queryKeys.subEvents(eventId),
            updater: (created) => appendToArray(created),
            showToast: { success: MSG_SUBEVENT_CREATED },
        })
    );
}

export function useUpdateSubEventMutation(eventId: string, subEventId: string) {
    return useWithExecute(
        useArrayMutation<SubEvent, SubEvent, UpdateSubEventInput>({
            mutationFn: (data) =>
                apiRequest(
                    {
                        url: `/events/${eventId}/subevents/${subEventId}`,
                        method: "PATCH",
                        data,
                    },
                    subEventResponseSchema
                ),
            queryKey: queryKeys.subEvents(eventId),
            updater: (updated) => replaceInArray(updated),
            showToast: { success: MSG_SUBEVENT_UPDATED },
        })
    );
}

export function useDeleteSubEventMutation(eventId: string) {
    return useWithExecute(
        useArrayMutation<SubEvent, void, string>({
            mutationFn: (subEventId) =>
                apiRequest(
                    {
                        url: `/events/${eventId}/subevents/${subEventId}`,
                        method: "DELETE",
                    },
                    z.any()
                ),
            queryKey: queryKeys.subEvents(eventId),
            updater: (_res, subEventId) => removeFromArray(subEventId),
            showToast: { success: MSG_SUBEVENT_DELETED },
        })
    );
}

