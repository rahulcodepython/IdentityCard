import { InfiniteData, useInfiniteQuery, useQuery } from "@tanstack/react-query";

import { apiRequest } from "@/react-query/client";
import { queryKeys } from "@/react-query/query-keys";
import {
    ApplicantFilter,
    FormSummary,
    FormSummarySchema,
    PaginatedApplicantResponse,
    PaginatedApplicantResponseSchema,
} from "@/schema/applicants.types";

export function useApplicantSchemaQuery(eventId: string) {
    return useQuery<FormSummary | null, Error>({
        queryKey: queryKeys.applicants.schema(eventId),
        queryFn: async () => {
            return apiRequest<FormSummary | null>(
                {
                    url: `/events/${eventId}/applicants/schema`,
                    method: "GET",
                },
                FormSummarySchema.nullable(),
            );
        },
        enabled: Boolean(eventId),
        staleTime: 5 * 60 * 1000,
    });
}

export function useApplicantsInfiniteQuery(
    eventId: string,
    filters?: { search?: string; filters?: ApplicantFilter[] },
) {
    const search = filters?.search?.trim() || undefined;
    const dynamicFilters =
        filters?.filters && filters.filters.length > 0
            ? filters.filters
            : undefined;

    return useInfiniteQuery<
        PaginatedApplicantResponse,
        Error,
        InfiniteData<PaginatedApplicantResponse>,
        readonly unknown[],
        number
    >({
        queryKey: queryKeys.applicants.byEvent(eventId, {
            search,
            filters: dynamicFilters,
        }),
        queryFn: async ({ pageParam = 1 }) => {
            return apiRequest<PaginatedApplicantResponse>(
                {
                    url: `/events/${eventId}/applicants`,
                    method: "GET",
                    params: {
                        page: pageParam,
                        limit: 30,
                        search,
                        filters: dynamicFilters
                            ? JSON.stringify(dynamicFilters)
                            : undefined,
                    },
                },
                PaginatedApplicantResponseSchema,
            );
        },
        initialPageParam: 1,
        getNextPageParam: (lastPage) => {
            if (!lastPage || lastPage.data.length < lastPage.limit) {
                return undefined;
            }
            if (lastPage.page * lastPage.limit >= lastPage.total) {
                return undefined;
            }
            return lastPage.page + 1;
        },
        enabled: Boolean(eventId),
    });
}
