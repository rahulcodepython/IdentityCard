"use client";

import { apiRequest } from "@/react-query/client";
import {
    useSimpleMutation,
    useWithExecute,
} from "@/react-query/mutation";
import { useAppQuery } from "@/react-query/query";
import { queryKeys } from "@/react-query/query-keys";
import {
    type OrgBilling,
    type Plan,
    type PurchaseInput,
    type UpgradeInput,
    billingResponseSchema,
    orgBillingResponseSchema,
    plansResponseSchema,
} from "@/schema/plans.types";

export function useListPlansQuery(enabled = true) {
    return useAppQuery<Plan[]>(
        queryKeys.plans(),
        () =>
            apiRequest(
                { url: "/plans", method: "GET" },
                plansResponseSchema
            ),
        { enabled }
    );
}

export function useListBillingQuery(enabled = true) {
    return useAppQuery<OrgBilling>(
        queryKeys.orgSubscriptions(),
        () =>
            apiRequest(
                { url: "/plans/billing", method: "GET" },
                orgBillingResponseSchema
            ),
        { enabled }
    );
}

// Alias for backward compatibility
export const useListSubscriptionsQuery = useListBillingQuery;

export function usePurchasePlanMutation() {
    return useWithExecute(
        useSimpleMutation<unknown, PurchaseInput>({
            mutationFn: (data) =>
                apiRequest(
                    {
                        url: "/plans/purchase",
                        method: "POST",
                        data,
                    },
                    billingResponseSchema
                ),
            invalidateKeys: [queryKeys.orgSubscriptions(), queryKeys.plans()],
            showToast: { success: "Plan purchased successfully" },
        })
    );
}

export function useRenewPlanMutation() {
    return useWithExecute(
        useSimpleMutation<unknown, { lineageRootId: string }>({
            mutationFn: ({ lineageRootId }) =>
                apiRequest(
                    {
                        url: `/plans/${lineageRootId}/renew`,
                        method: "POST",
                    },
                    billingResponseSchema
                ),
            invalidateKeys: [queryKeys.orgSubscriptions()],
            showToast: { success: "Plan renewed successfully" },
        })
    );
}

export function useUpgradePlanMutation() {
    return useWithExecute(
        useSimpleMutation<unknown, { lineageRootId: string; data: UpgradeInput }>({
            mutationFn: ({ lineageRootId, data }) =>
                apiRequest(
                    {
                        url: `/plans/${lineageRootId}/upgrade`,
                        method: "POST",
                        data,
                    },
                    billingResponseSchema
                ),
            invalidateKeys: [queryKeys.orgSubscriptions()],
            showToast: { success: "Plan upgraded successfully" },
        })
    );
}

export function useCancelPlanMutation() {
    return useWithExecute(
        useSimpleMutation<unknown, { lineageRootId: string }>({
            mutationFn: ({ lineageRootId }) =>
                apiRequest(
                    {
                        url: `/plans/${lineageRootId}/cancel`,
                        method: "POST",
                    },
                    billingResponseSchema
                ),
            invalidateKeys: [queryKeys.orgSubscriptions()],
            showToast: { success: "Plan cancelled" },
        })
    );
}
