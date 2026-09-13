import {
    type RemixiconComponentType,
    RiBankCardLine,
    RiCalendarEventLine,
    RiDashboardLine,
    RiSettings3Line,
    RiSmartphoneLine,
    RiUser3Line,
} from "@remixicon/react";

export type NavItem = {
    title: string;
    url?: string;
    icon: RemixiconComponentType;
    items?: {
        title: string;
        url: string;
        icon: RemixiconComponentType;
    }[];
};

export type NavGroup = {
    label: string;
    items: NavItem[];
};

const getNavItems = (isOwner: boolean, orgSlug: string): NavGroup[] => {
    const prefix = `/dashboard/${orgSlug}`;
    const items: NavGroup[] = [
        {
            label: "Dashboard",
            items: [{ title: "Dashboard", url: prefix, icon: RiDashboardLine }],
        },
        {
            label: "Events Management",
            items: [
                {
                    title: "Events",
                    url: `${prefix}/events`,
                    icon: RiCalendarEventLine,
                },
                {
                    title: "Members",
                    url: `${prefix}/members`,
                    icon: RiUser3Line,
                },
                {
                    title: "Devices",
                    url: `${prefix}/devices`,
                    icon: RiSmartphoneLine,
                },
            ],
        },
    ];

    if (isOwner) {
        items.push({
            label: "Management",
            items: [
                {
                    title: "Billing",
                    url: `${prefix}/billing`,
                    icon: RiBankCardLine,
                },
                {
                    title: "Settings",
                    url: `${prefix}/settings`,
                    icon: RiSettings3Line,
                },
            ],
        });
    }

    return items;
};

export default getNavItems;