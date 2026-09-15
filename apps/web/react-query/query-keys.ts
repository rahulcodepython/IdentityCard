export const queryKeys = {
    events: {
        all: ["events"] as const,
        list: (filters?: { search?: string }) => ["events", "list", filters] as const,
        detail: (id: string) => ["events", "detail", id] as const,
    },
    eventDates: {
        all: ["eventDates"] as const,
        byEvent: (eventId: string) => ["eventDates", "event", eventId] as const,
        byMonth: (eventId: string, month: string) =>
            ["eventDates", "event", eventId, "month", month] as const,
    },
    forms: {
        all: ["forms"] as const,
        list: (filters?: { search?: string }) => ["forms", "list", filters] as const,
        detail: (id: string) => ["forms", "detail", id] as const,
    },
} as const;
