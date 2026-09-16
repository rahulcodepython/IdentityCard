import * as React from "react";
import { Calendar, Smartphone } from "lucide-react";

import { DeviceActionsMenu } from "@/components/devices/device-actions-menu";
import { DevicePINDisplay } from "@/components/devices/device-pin-display";
import { DeviceStatusBadge } from "@/components/devices/device-status-badge";
import type { Device } from "@/schema/devices.types";

interface GlobalDevicesTableProps {
    devices: Device[];
}

export function GlobalDevicesTable({ devices }: GlobalDevicesTableProps) {
    const formatTTL = (expiresAt?: string | null) => {
        if (!expiresAt) return "Unlimited";
        try {
            const d = new Date(expiresAt);
            return isNaN(d.getTime()) ? "Unlimited" : d.toLocaleDateString(undefined, {
                month: "short",
                day: "numeric",
                year: "numeric",
            });
        } catch {
            return "Unlimited";
        }
    };

    return (
        <div className="rounded-xl border bg-card shadow-xs overflow-hidden">
            <div className="overflow-x-auto">
                <table className="w-full text-left text-xs">
                    <thead className="bg-muted/40 border-b text-muted-foreground font-medium">
                        <tr>
                            <th className="py-2.5 px-4">Terminal Name</th>
                            <th className="py-2.5 px-4">Status</th>
                            <th className="py-2.5 px-4">Fingerprint</th>
                            <th className="py-2.5 px-4">Expiration (TTL)</th>
                            <th className="py-2.5 px-4">PIN / Pair Key</th>
                            <th className="py-2.5 px-4 text-right">Actions</th>
                        </tr>
                    </thead>
                    <tbody className="divide-y divide-border">
                        {devices.map((device) => (
                            <tr key={device.id} className="hover:bg-muted/25 transition-colors">
                                <td className="py-3 px-4">
                                    <div className="flex items-center gap-2.5">
                                        <div className="flex size-7 items-center justify-center rounded-lg bg-primary/10 text-primary">
                                            <Smartphone className="size-3.5" />
                                        </div>
                                        <div>
                                            <p className="font-semibold text-foreground text-xs">{device.name}</p>
                                            <p className="text-[11px] text-muted-foreground">
                                                {device.actual_name || "Hardware unverified"}
                                            </p>
                                        </div>
                                    </div>
                                </td>
                                <td className="py-3 px-4 whitespace-nowrap">
                                    <DeviceStatusBadge device={device} />
                                </td>
                                <td className="py-3 px-4 whitespace-nowrap font-mono text-[11px] text-muted-foreground">
                                    {device.fingerprint ? (
                                        <span className="truncate block max-w-[140px]" title={device.fingerprint}>
                                            {device.fingerprint}
                                        </span>
                                    ) : (
                                        <span className="italic text-[11px]">None</span>
                                    )}
                                </td>
                                <td className="py-3 px-4 whitespace-nowrap text-muted-foreground text-[11px]">
                                    <div className="flex items-center gap-1.5">
                                        <Calendar className="size-3 text-muted-foreground" />
                                        <span>{formatTTL(device.expires_at)}</span>
                                    </div>
                                </td>
                                <td className="py-3 px-4 whitespace-nowrap">
                                    <DevicePINDisplay device={device} />
                                </td>
                                <td className="py-3 px-4 text-right whitespace-nowrap">
                                    <DeviceActionsMenu device={device} />
                                </td>
                            </tr>
                        ))}
                    </tbody>
                </table>
            </div>
        </div>
    );
}
