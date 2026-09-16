import * as React from "react";
import { Smartphone } from "lucide-react";

import { DeviceActionsMenu } from "@/components/devices/device-actions-menu";
import { DeviceStatusBadge } from "@/components/devices/device-status-badge";
import type { EventDeviceAssignment } from "@/schema/devices.types";

interface EventDevicesTableProps {
    assignments: EventDeviceAssignment[];
    eventId: string;
}

export function EventDevicesTable({ assignments, eventId }: EventDevicesTableProps) {
    return (
        <div className="rounded-xl border bg-card shadow-xs overflow-hidden">
            <div className="overflow-x-auto">
                <table className="w-full text-left text-xs">
                    <thead className="bg-muted/40 border-b text-muted-foreground font-medium">
                        <tr>
                            <th className="py-2.5 px-4">Scanner Terminal</th>
                            <th className="py-2.5 px-4">Status</th>
                            <th className="py-2.5 px-4">Fingerprint</th>
                            <th className="py-2.5 px-4">Assigned On</th>
                            <th className="py-2.5 px-4 text-right">Actions</th>
                        </tr>
                    </thead>
                    <tbody className="divide-y divide-border">
                        {assignments.map((assignment) => {
                            const device = assignment.device;
                            return (
                                <tr key={assignment.id} className="hover:bg-muted/25 transition-colors">
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
                                            <span className="italic text-[11px]">Unpaired</span>
                                        )}
                                    </td>
                                    <td className="py-3 px-4 whitespace-nowrap text-muted-foreground text-[11px]">
                                        {new Date(assignment.created_at).toLocaleDateString()}
                                    </td>
                                    <td className="py-3 px-4 text-right whitespace-nowrap">
                                        <DeviceActionsMenu device={device} eventId={eventId} />
                                    </td>
                                </tr>
                            );
                        })}
                    </tbody>
                </table>
            </div>
        </div>
    );
}
