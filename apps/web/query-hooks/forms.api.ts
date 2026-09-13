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
import { queryKeys } from "@/react-query/query-keys";
import {
    type CreateFormInput,
    type EventForm,
    type PublicForm,
    type SubmitFormInput,
    type UpdateFormInput,
    formResponseSchema,
    formsListResponseSchema,
    messageResponseSchema,
    publicFormResponseSchema,
} from "@/schema/forms.types";
import {
    MSG_FORM_DELETED,
    MSG_FORM_LINK_CREATED,
    MSG_FORM_SETTINGS_UPDATED,
    MSG_REGISTRATION_SUBMITTED,
} from "@/lib/constants";
import { z } from "zod";

export function useFormsListQuery(eventId: string, enabled = true) {
    return useAppQuery<EventForm[]>(
        queryKeys.forms(eventId),
        () =>
            apiRequest(
                { url: `/events/${eventId}/forms`, method: "GET" },
                formsListResponseSchema
            ),
        { enabled: enabled && !!eventId }
    );
}

export function usePublicFormQuery(token: string, enabled = true) {
    return useAppQuery<PublicForm>(
        queryKeys.publicForm(token),
        () =>
            apiRequest(
                { url: `/public/forms/${token}`, method: "GET" },
                publicFormResponseSchema
            ),
        { enabled: enabled && !!token }
    );
}

export function useCreateFormMutation(eventId: string) {
    return useWithExecute(
        useArrayMutation<EventForm, EventForm, CreateFormInput>({
            mutationFn: (data) =>
                apiRequest(
                    {
                        url: `/events/${eventId}/forms`,
                        method: "POST",
                        data,
                    },
                    formResponseSchema
                ),
            queryKey: queryKeys.forms(eventId),
            updater: (created) => appendToArray(created),
            showToast: { success: MSG_FORM_LINK_CREATED },
        })
    );
}

export function useUpdateFormMutation(eventId: string, formId: string) {
    return useWithExecute(
        useArrayMutation<EventForm, EventForm, UpdateFormInput>({
            mutationFn: (data) =>
                apiRequest(
                    {
                        url: `/events/${eventId}/forms/${formId}`,
                        method: "PATCH",
                        data,
                    },
                    formResponseSchema
                ),
            queryKey: queryKeys.forms(eventId),
            updater: (updated) => replaceInArray(updated),
            showToast: { success: MSG_FORM_SETTINGS_UPDATED },
        })
    );
}

export function useDeleteFormMutation(eventId: string) {
    return useWithExecute(
        useArrayMutation<EventForm, void, string>({
            mutationFn: (formId) =>
                apiRequest(
                    {
                        url: `/events/${eventId}/forms/${formId}`,
                        method: "DELETE",
                    },
                    z.any()
                ),
            queryKey: queryKeys.forms(eventId),
            updater: (_res, formId) => removeFromArray(formId),
            showToast: { success: MSG_FORM_DELETED },
        })
    );
}

export function useSubmitPublicFormMutation(token: string) {
    return useWithExecute(
        useSimpleMutation<{ message: string }, SubmitFormInput>({
            mutationFn: (data) =>
                apiRequest(
                    {
                        url: `/public/forms/${token}/submit`,
                        method: "POST",
                        data,
                    },
                    messageResponseSchema
                ),
            invalidateKeys: [queryKeys.publicForm(token)],
            showToast: { success: MSG_REGISTRATION_SUBMITTED },
        })
    );
}

