"use client";

import * as React from "react";
import { useParams } from "next/navigation";
import { Loader2 } from "lucide-react";

import { SharingConfigForm } from "@/components/events/sharing/sharing-config-form";
import { SharingHeader } from "@/components/events/sharing/sharing-header";
import { SharingMetricsCard } from "@/components/events/sharing/sharing-metrics-card";
import { SharingQrCard } from "@/components/events/sharing/sharing-qr-card";
import { useBreadcrumbs } from "@/hooks/use-breadcrumbs";
import { useEventQuery } from "@/query-hooks/events.api";
import { useEventSharingQuery } from "@/query-hooks/eventsharing.api";

export default function EventSharingPage() {
    const params = useParams();
    const eventId = params?.eventId as string;

    const { data: event } = useEventQuery(eventId);
    const { data: sharing, isLoading: isSharingLoading } = useEventSharingQuery(eventId);

    useBreadcrumbs([
        {
            title: "Dashboard",
            url: "/dashboard",
        },
        {
            title: "Events",
            url: "/dashboard/events",
        },
        {
            title: event?.name || "Event",
            url: `/dashboard/events/${eventId}`,
        },
        {
            title: "Sharing",
        },
    ]);

    if (isSharingLoading) {
        return (
            <div className="flex h-64 w-full items-center justify-center">
                <div className="flex flex-col items-center gap-2">
                    <Loader2 className="size-6 animate-spin text-primary" />
                    <span className="text-xs text-muted-foreground">Loading sharing configuration...</span>
                </div>
            </div>
        );
    }

    const isLive = sharing?.event_form?.status === "live";

    return (
        <div className="flex flex-col gap-5 max-w-6xl">
            <SharingHeader eventName={event?.name} isLive={isLive} />

            <div className="grid grid-cols-1 lg:grid-cols-3 gap-5 items-start">
                {/* Left 2 columns: Configuration Form */}
                <div className="lg:col-span-2">
                    <SharingConfigForm eventId={eventId} sharing={sharing} />
                </div>

                {/* Right 1 column: QR Code & Status Overview */}
                <div className="flex flex-col gap-5">
                    <SharingQrCard
                        eventFormId={sharing?.event_form?.id}
                        isLive={isLive}
                    />
                    <SharingMetricsCard sharing={sharing} />
                </div>
            </div>
        </div>
    );
}
