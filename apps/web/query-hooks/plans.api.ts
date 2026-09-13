"use client";

import { apiRequest } from "@/react-query/client";
import {
    useSimpleMutation,
    useWithExecute,
} from "@/react-query/mutation";
import { useAppQuery } from "@/react-query/query";
import { queryKeys } from "@/react-query/query-keys";
import {
    type BillingOverview,
    type PurchaseCreditsInput,
    billingOverviewResponseSchema,
} from "@/schema/plans.types";
import { MSG_ANNUAL_RENEWED, MSG_CREDITS_PURCHASED } from "@/lib/constants";

export function useBillingOverviewQuery(enabled = true) {
    return useAppQuery<BillingOverview>(
        queryKeys.orgSubscriptions(),
        () =>
            apiRequest(
                { url: "/billing", method: "GET" },
                billingOverviewResponseSchema
            ),
        { enabled }
    );
}

export const useListBillingQuery = useBillingOverviewQuery;

export function usePurchaseCreditsMutation() {
    return useWithExecute(
        useSimpleMutation<BillingOverview, PurchaseCreditsInput>({
            mutationFn: (data) =>
                apiRequest(
                    {
                        url: "/billing/credits/purchase",
                        method: "POST",
                        data,
                    },
                    billingOverviewResponseSchema
                ),
            invalidateKeys: [queryKeys.orgSubscriptions()],
            showToast: { success: MSG_CREDITS_PURCHASED },
        })
    );
}

export function useRenewAnnualMutation() {
    return useWithExecute(
        useSimpleMutation<BillingOverview, void>({
            mutationFn: () =>
                apiRequest(
                    {
                        url: "/billing/renew",
                        method: "POST",
                    },
                    billingOverviewResponseSchema
                ),
            invalidateKeys: [queryKeys.orgSubscriptions()],
            showToast: { success: MSG_ANNUAL_RENEWED },
        })
    );
}
