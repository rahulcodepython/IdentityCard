import useUser from "@/hooks/use-user";
import { authClient } from "@/lib/auth-client";
import { useOrganizationsQuery } from "@/query-hooks/organizations.api";
import { ListOrganizationsItem } from "@/schema/organizations.types";
import { ROUTE_DASHBOARD } from "@/lib/constants";
import { useSessionStore } from "@/store/session.store";
import { usePathname, useRouter } from "next/navigation";
import React, { useEffect, useState } from "react";

const useOrganization = (orgSlug: string | null) => {
    const { data: organizations, isLoading } = useOrganizationsQuery();

    const [isFetched, setIsFetched] = useState(false);

    const [currentOrg, setCurrentOrg] =
        React.useState<ListOrganizationsItem | null>(null);

    const router = useRouter();
    const pathname = usePathname();

    const { activeOrgId } = useUser();

    const redirectToDashboard = () => {
        if (pathname !== ROUTE_DASHBOARD) return router.push(ROUTE_DASHBOARD);
        return;
    };

    const handleSelectOrg = async (orgSlug: string | null) => {
        if (!organizations) return;

        if (!orgSlug) {
            if (!activeOrgId) return redirectToDashboard();

            const currentOrg = organizations.find(
                (org) => org.id === activeOrgId,
            );

            if (!currentOrg) return redirectToDashboard();

            setCurrentOrg(currentOrg);
            return router.push(`/dashboard/${currentOrg.slug}`);
        }

        const currentOrg = organizations.find((org) => org.slug === orgSlug);

        if (!currentOrg) return redirectToDashboard();

        await authClient.organization
            .setActive({
                organizationId: currentOrg.id,
            })
            .then(() => {
                setCurrentOrg(currentOrg);

                orgSlug === currentOrg.slug
                    ? null
                    : router.push(`/dashboard/${orgSlug}`);
            })
            .catch((err) => {
                console.error("Failed to set active organization:", err);
            })
            .finally(() => {
                setIsFetched(true);
            });
    };

    useEffect(() => {
        if (isLoading || isFetched) return;

        handleSelectOrg(orgSlug);
    }, [organizations, router, isLoading, isFetched, orgSlug]);

    const isOwner = currentOrg?.role === "owner";

    return {
        organizations,
        currentOrg,
        isLoading,
        handleSelectOrg,
        isOwner,
        isAdmin: isOwner,
    };
};

export default useOrganization;
