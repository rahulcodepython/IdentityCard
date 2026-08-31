"use client";

import { apiRequest } from "@/react-query/client";
import { useAppQuery } from "@/react-query/query";
import { queryKeys } from "@/react-query/query-keys";
import {
    type AnalyticsOverview,
    type AnalyticsSummary,
    type DailyBreakdown,
    dailyResponseSchema,
    overviewResponseSchema,
    summaryResponseSchema,
} from "@/schema/analytics.types";

export function useAnalyticsSummaryQuery(eventId: string, enabled = true) {
    return useAppQuery<AnalyticsSummary>(
        queryKeys.analyticsSummary(eventId),
        () =>
            apiRequest(
                { url: `/events/${eventId}/analytics/summary`, method: "GET" },
                summaryResponseSchema
            ),
        { enabled: enabled && !!eventId }
    );
}

export function useAnalyticsDailyQuery(eventId: string, subEventId?: string, enabled = true) {
    const qs = subEventId ? `?sub_event_id=${subEventId}` : "";
    return useAppQuery<DailyBreakdown[]>(
        queryKeys.analyticsDaily(eventId, subEventId),
        async () => {
            const res = await apiRequest(
                { url: `/events/${eventId}/analytics/daily${qs}`, method: "GET" },
                dailyResponseSchema
            );
            return res.days;
        },
        { enabled: enabled && !!eventId }
    );
}

export function useAnalyticsOverviewQuery(enabled = true) {
    return useAppQuery<AnalyticsOverview>(
        queryKeys.analyticsOverview(),
        () =>
            apiRequest(
                { url: "/analytics/overview", method: "GET" },
                overviewResponseSchema
            ),
        { enabled }
    );
}

