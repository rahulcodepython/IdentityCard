export type PeopleFilter = { subEventId?: string; search?: string };
export type RosterFilter = {
    subEventId?: string;
    date?: string;
    attended?: boolean;
    status?: string;
};

export const queryKeys = {
    // auth / session
    me: () => ["me"] as const,

    // organizations
    orgSettings: () => ["organizations", "settings"] as const,

    // plans / billing
    plans: () => ["plans"] as const,
    orgSubscriptions: () => ["plans", "subscriptions"] as const,

    // members
    members: () => ["members"] as const,

    // devices
    devices: () => ["devices"] as const,

    // events
    events: () => ["events"] as const,
    event: (eventId: string) => ["events", eventId] as const,

    // sub-events (nested under an event)
    subEvents: (eventId: string) => ["events", eventId, "subevents"] as const,
    subEvent: (eventId: string, id: string) =>
        ["events", eventId, "subevents", id] as const,

    // people (nested under an event)
    people: (eventId: string, filter: PeopleFilter = {}) =>
        ["events", eventId, "people", filter] as const,
    person: (eventId: string, id: string) =>
        ["events", eventId, "people", id] as const,

    // forms (nested under an event) — plus the public, unauthenticated read
    forms: (eventId: string) => ["events", eventId, "forms"] as const,
    publicForm: (token: string) => ["public", "forms", token] as const,

    // attendance roster (nested under an event)
    roster: (eventId: string, filter: RosterFilter = {}) =>
        ["events", eventId, "attendance", filter] as const,

    // analytics (nested under an event, plus org-wide overview)
    analyticsSummary: (eventId: string) =>
        ["events", eventId, "analytics", "summary"] as const,
    analyticsDaily: (eventId: string, subEventId?: string) =>
        ["events", eventId, "analytics", "daily", subEventId ?? null] as const,
    analyticsOverview: () => ["analytics", "overview"] as const,

    // scanner device (device-key auth, not JWT — see store/device.store.ts)
    scannerMe: () => ["scanner", "me"] as const,
} as const;
