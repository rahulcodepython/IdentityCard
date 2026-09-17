import { useMutation, useQuery } from "@tanstack/react-query";
import { toast } from "sonner";

import { apiRequest } from "../react-query/client";
import { queryKeys } from "../react-query/query-keys";
import {
    PublicApplyConfig,
    PublicApplyConfigSchema,
    SubmitApplicationInput,
    SubmitApplicationResponse,
    SubmitApplicationResponseSchema,
    SubmitApplicationSchema,
} from "../schema/publicapply.types";

export function usePublicApplyConfigQuery(eventFormId: string) {
    return useQuery<PublicApplyConfig, Error>({
        queryKey: queryKeys.publicApply.detail(eventFormId),
        queryFn: async () => {
            return apiRequest<PublicApplyConfig>(
                {
                    url: `/public/apply/${eventFormId}`,
                    method: "GET",
                },
                PublicApplyConfigSchema,
            );
        },
        enabled: Boolean(eventFormId),
        retry: false,
    });
}

export function useSubmitApplicationMutation() {
    return useMutation<
        SubmitApplicationResponse,
        Error,
        { eventFormId: string } & SubmitApplicationInput
    >({
        mutationFn: async ({ eventFormId, ...data }) => {
            const validated = SubmitApplicationSchema.parse(data);
            return apiRequest<SubmitApplicationResponse>(
                {
                    url: `/public/apply/${eventFormId}`,
                    method: "POST",
                    data: validated,
                },
                SubmitApplicationResponseSchema,
            );
        },
        onError: (error) => {
            toast.error(error.message || "Failed to submit application");
        },
    });
}
