"use client";

import { authClient } from "@/lib/auth-client";
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
    type ListOrganizationsItem,
    listOrganizationsResponseSchema,
} from "@/schema/organizations.types";
import {
    MSG_LOGO_REMOVED,
    MSG_LOGO_UPLOADED,
    MSG_ORG_DELETED,
    MSG_ORG_SETTINGS_UPDATED,
} from "@/lib/constants";
import { useSessionStore } from "@/store/session.store";
import { z } from "zod";

export function useOrganizationsQuery(enabled = true) {
    return useAppQuery<ListOrganizationsItem[]>(
        queryKeys.organizations(),
        () => apiRequest(
            { url: "/organizations", method: "GET" },
            listOrganizationsResponseSchema
        ),
        { enabled }
    );
}


export function useOrgSettingsQuery(activeOrgId?: string, enabled = true) {
    const orgId = activeOrgId || useSessionStore.getState().activeOrgId || undefined;
    return useAppQuery<OrgSettings>(
        queryKeys.orgSettings(),
        () => apiRequest(
            {
                url: orgId ? `/organization/${orgId}/settings` : "/organizations/settings",
                method: "GET",
            },
            settingsResponseSchema
        ),
        { enabled: enabled && !!orgId }
    );
}

export function useUpdateOrgSettingsMutation(activeOrgId?: string) {
    return useWithExecute(
        useObjectMutation<OrgSettings, { organization_name: string }>({
            mutationFn: (data) => {
                const orgId = activeOrgId || useSessionStore.getState().activeOrgId || undefined;
                return apiRequest(
                    {
                        url: orgId ? `/organization/${orgId}/settings` : "/organizations/settings",
                        method: "PATCH",
                        data,
                    },
                    settingsResponseSchema
                );
            },
            queryKey: queryKeys.orgSettings(),
            updater: (updated) => () => updated,
            showToast: { success: MSG_ORG_SETTINGS_UPDATED },
        })
    );
}

export function useDeleteOrganizationMutation(activeOrgId?: string) {
    return useWithExecute(
        useSimpleMutation<void, string | void>({
            mutationFn: (orgId?: string | void) => {
                const targetId = (typeof orgId === "string" ? orgId : "") || activeOrgId || useSessionStore.getState().activeOrgId || "";
                return apiRequest(
                    {
                        url: targetId ? `/organization/${targetId}` : "/organizations",
                        method: "DELETE",
                    },
                    z.any()
                );
            },
            invalidateKeys: [queryKeys.organizations(), queryKeys.orgSettings()],
            showToast: { success: MSG_ORG_DELETED },
        })
    );
}



export function useUploadLogoMutation() {
    return useWithExecute(
        useSimpleMutation<void, FormData>({
            mutationFn: (formData) => apiRequest(
                {
                    url: "/organizations/settings/logo",
                    method: "POST",
                    data: formData,
                    headers: { "Content-Type": "multipart/form-data" },
                },
                z.any()
            ),
            invalidateKeys: [queryKeys.orgSettings()],
            showToast: { success: MSG_LOGO_UPLOADED },
        })
    );
}

export function useDeleteLogoMutation() {
    return useWithExecute(
        useSimpleMutation({
            mutationFn: () => apiRequest(
                {
                    url: "/organizations/settings/logo",
                    method: "DELETE",
                },
                z.any()
            ),
            invalidateKeys: [queryKeys.orgSettings()],
            showToast: { success: MSG_LOGO_REMOVED },
        })
    );
}

