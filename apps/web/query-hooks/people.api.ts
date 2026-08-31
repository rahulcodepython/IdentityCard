"use client";

import { apiRequest } from "@/react-query/client";
import {
    appendToArray,
    removeFromArray,
    replaceInArray,
    useArrayMutation,
    useSimpleMutation,
    useWithExecute,
} from "@/react-query/mutation";
import { useAppQuery } from "@/react-query/query";
import { type PeopleFilter, queryKeys } from "@/react-query/query-keys";
import {
    type CreatePersonInput,
    type ImportSummary,
    type Person,
    type UpdatePersonInput,
    importSummarySchema,
    peopleListResponseSchema,
    personResponseSchema,
} from "@/schema/people.types";
import { z } from "zod";

export function usePeopleListQuery(eventId: string, filter: PeopleFilter = {}, enabled = true) {
    const params = new URLSearchParams();
    if (filter.subEventId) params.set("sub_event_id", filter.subEventId);
    if (filter.search) params.set("search", filter.search);
    const qs = params.toString();

    return useAppQuery<Person[]>(
        queryKeys.people(eventId, filter),
        () =>
            apiRequest(
                { url: `/events/${eventId}/people${qs ? `?${qs}` : ""}`, method: "GET" },
                peopleListResponseSchema
            ),
        { enabled: enabled && !!eventId }
    );
}

export function usePersonDetailQuery(eventId: string, personId: string, enabled = true) {
    return useAppQuery<Person>(
        queryKeys.person(eventId, personId),
        () =>
            apiRequest(
                { url: `/events/${eventId}/people/${personId}`, method: "GET" },
                personResponseSchema
            ),
        { enabled: enabled && !!eventId && !!personId }
    );
}

export function useCreatePersonMutation(eventId: string, filter: PeopleFilter = {}) {
    return useWithExecute(
        useArrayMutation<Person, Person, CreatePersonInput>({
            mutationFn: (data) =>
                apiRequest(
                    {
                        url: `/events/${eventId}/people`,
                        method: "POST",
                        data,
                    },
                    personResponseSchema
                ),
            queryKey: queryKeys.people(eventId, filter),
            updater: (created) => appendToArray(created),
            invalidateKeys: [queryKeys.analyticsSummary(eventId)],
            showToast: { success: "Attendee added" },
        })
    );
}

export function useUpdatePersonMutation(eventId: string, personId: string, filter: PeopleFilter = {}) {
    return useWithExecute(
        useArrayMutation<Person, Person, UpdatePersonInput>({
            mutationFn: (data) =>
                apiRequest(
                    {
                        url: `/events/${eventId}/people/${personId}`,
                        method: "PATCH",
                        data,
                    },
                    personResponseSchema
                ),
            queryKey: queryKeys.people(eventId, filter),
            updater: (updated) => replaceInArray(updated),
            invalidateKeys: [queryKeys.person(eventId, personId)],
            showToast: { success: "Attendee updated" },
        })
    );
}

export function useDeletePersonMutation(eventId: string, filter: PeopleFilter = {}) {
    return useWithExecute(
        useArrayMutation<Person, void, string>({
            mutationFn: (personId) =>
                apiRequest(
                    {
                        url: `/events/${eventId}/people/${personId}`,
                        method: "DELETE",
                    },
                    z.any()
                ),
            queryKey: queryKeys.people(eventId, filter),
            updater: (_res, personId) => removeFromArray(personId),
            invalidateKeys: [queryKeys.analyticsSummary(eventId)],
            showToast: { success: "Attendee removed" },
        })
    );
}

export function useImportPeopleMutation(eventId: string) {
    return useWithExecute(
        useSimpleMutation<ImportSummary, { formData: FormData; subEventId?: string }>({
            mutationFn: ({ formData, subEventId }) => {
                if (subEventId) {
                    formData.set("sub_event_id", subEventId);
                }
                return apiRequest(
                    {
                        url: `/events/${eventId}/people/import`,
                        method: "POST",
                        data: formData,
                        headers: { "Content-Type": "multipart/form-data" },
                    },
                    importSummarySchema
                );
            },
            invalidateKeys: [queryKeys.people(eventId), queryKeys.analyticsSummary(eventId)],
            showToast: { success: "Attendees CSV imported" },
        })
    );
}

