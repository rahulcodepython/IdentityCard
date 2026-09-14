import { LucideIcon } from "lucide-react";
import type * as React from "react";

// Sub-item in collapsible navigation
export interface NavSubItem {
    title: string;
    url: string;
}

// Main navigation item with optional sub-items and icon
export interface NavNestedItem {
    title: string;
    url: string;
    icon?: LucideIcon;
    items?: NavSubItem[];
}

// Props for NavMain component
export interface NavNestedItems {
    label: string;
    items: NavNestedItem[];
}

// Project navigation item
export interface NavItem {
    name: string;
    url: string;
    icon: LucideIcon;
}

// Props for NavProjects component
export interface NavItems {
    label: string;
    items: NavItem[];
}

// Sidebar user profile information
export interface SidebarUser {
    id: string;
    name: string;
    email: string;
    avatar: string;
}

// Props for NavUser component
export interface NavUserProps {
    user: SidebarUser;
}

// Team or organization item in the switcher
export interface SidebarTeam {
    name: string;
    logo: React.ElementType;
    plan: string;
}

// Props for TeamSwitcher component
export interface TeamSwitcherProps {
    teams: SidebarTeam[];
}

// Complete sidebar data model
export interface SidebarData {
    user: SidebarUser;
    teams: SidebarTeam[];
    dashboard: NavItems;
    management: NavItems;
}
