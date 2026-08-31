"use client";

import { apiRequest } from "@/react-query/client";
import {
    useObjectMutation,
    useSimpleMutation,
    useWithExecute,
} from "@/react-query/mutation";
import { useAppQuery } from "@/react-query/query";
import { queryKeys } from "@/react-query/query-keys";
import {
    type OrgSettings,
    settingsResponseSchema,
} from "@/schema/organizations.types";
import { z } from "zod";

export function useOrgSettingsQuery(enabled = true) {
    return useAppQuery<OrgSettings>(
        queryKeys.orgSettings(),
        () =>
            apiRequest(
                { url: "/organizations/settings", method: "GET" },
                settingsResponseSchema
            ),
        { enabled }
    );
}

export function useUpdateOrgSettingsMutation() {
    return useWithExecute(
        useObjectMutation<OrgSettings, { organization_name: string }>({
            mutationFn: (data) =>
                apiRequest(
                    {
                        url: "/organizations/settings",
                        method: "PATCH",
                        data,
                    },
                    settingsResponseSchema
                ),
            queryKey: queryKeys.orgSettings(),
            updater: (updated) => () => updated,
            showToast: { success: "Organization settings updated" },
        })
    );
}

export function useDeleteOrganizationMutation() {
    return useWithExecute(
        useSimpleMutation({
            mutationFn: () =>
                apiRequest(
                    {
                        url: "/organizations",
                        method: "DELETE",
                    },
                    z.any()
                ),
            invalidateKeys: [queryKeys.orgSettings()],
            showToast: { success: "Organization deleted" },
        })
    );
}

export function useUploadLogoMutation() {
    return useWithExecute(
        useSimpleMutation<void, FormData>({
            mutationFn: (formData) =>
                apiRequest(
                    {
                        url: "/organizations/settings/logo",
                        method: "POST",
                        data: formData,
                        headers: { "Content-Type": "multipart/form-data" },
                    },
                    z.any()
                ),
            invalidateKeys: [queryKeys.orgSettings()],
            showToast: { success: "Logo uploaded" },
        })
    );
}

export function useDeleteLogoMutation() {
    return useWithExecute(
        useSimpleMutation({
            mutationFn: () =>
                apiRequest(
                    {
                        url: "/organizations/settings/logo",
                        method: "DELETE",
                    },
                    z.any()
                ),
            invalidateKeys: [queryKeys.orgSettings()],
            showToast: { success: "Logo removed" },
        })
    );
}

