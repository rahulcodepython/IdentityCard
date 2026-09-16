"use client";

import * as React from "react";
import {
    Calendar,
    Edit,
    ExternalLink,
    KeyRound,
    Loader2,
    MoreHorizontal,
    RefreshCw,
    Trash2,
    Unlink,
} from "lucide-react";

import {
    AlertDialog,
    AlertDialogAction,
    AlertDialogCancel,
    AlertDialogContent,
    AlertDialogDescription,
    AlertDialogFooter,
    AlertDialogHeader,
    AlertDialogTitle,
} from "@/components/ui/alert-dialog";
import { Button } from "@/components/ui/button";
import { Checkbox } from "@/components/ui/checkbox";
import {
    Dialog,
    DialogContent,
    DialogDescription,
    DialogFooter,
    DialogHeader,
    DialogTitle,
} from "@/components/ui/dialog";
import {
    DropdownMenu,
    DropdownMenuContent,
    DropdownMenuItem,
    DropdownMenuSeparator,
    DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
    useDeleteDeviceMutation,
    useRegeneratePINMutation,
    useUnassignEventDeviceMutation,
    useUpdateDeviceMutation,
} from "@/query-hooks/devices.api";
import type { Device } from "@/schema/devices.types";

interface DeviceActionsMenuProps {
    device: Device;
    eventId?: string;
}

export function DeviceActionsMenu({ device, eventId }: DeviceActionsMenuProps) {
    const [isEditOpen, setIsEditOpen] = React.useState(false);
    const [isDeleteOpen, setIsDeleteOpen] = React.useState(false);
    const [isUnassignOpen, setIsUnassignOpen] = React.useState(false);

    // Edit form states
    const [name, setName] = React.useState(device.name);
    const [isUnlimitedTTL, setIsUnlimitedTTL] = React.useState(device.expires_at == null);

    const formatForDatetimeLocal = (isoString?: string | null) => {
        if (!isoString) return "";
        try {
            const d = new Date(isoString);
            if (isNaN(d.getTime())) return "";
            const year = d.getFullYear();
            const month = String(d.getMonth() + 1).padStart(2, "0");
            const day = String(d.getDate()).padStart(2, "0");
            const hours = String(d.getHours()).padStart(2, "0");
            const minutes = String(d.getMinutes()).padStart(2, "0");
            return `${year}-${month}-${day}T${hours}:${minutes}`;
        } catch {
            return "";
        }
    };

    const [expiresAt, setExpiresAt] = React.useState(() => formatForDatetimeLocal(device.expires_at));

    const regenerateMutation = useRegeneratePINMutation();
    const updateMutation = useUpdateDeviceMutation();
    const deleteMutation = useDeleteDeviceMutation();
    const unassignMutation = useUnassignEventDeviceMutation();

    const handleSaveEdit = async (e: React.FormEvent) => {
        e.preventDefault();
        const trimmed = name.trim();
        if (!trimmed) return;

        const targetExpires = !isUnlimitedTTL && expiresAt ? new Date(expiresAt).toISOString() : null;

        await updateMutation.mutateAsync({
            id: device.id,
            values: { name: trimmed, expires_at: targetExpires },
        });
        setIsEditOpen(false);
    };

    const handleConfirmDelete = async () => {
        await deleteMutation.mutateAsync({ id: device.id });
        setIsDeleteOpen(false);
    };

    const handleConfirmUnassign = async () => {
        if (!eventId) return;
        await unassignMutation.mutateAsync({ eventId, deviceId: device.id });
        setIsUnassignOpen(false);
    };

    return (
        <>
            <DropdownMenu>
                <DropdownMenuTrigger
                    render={
                        <Button
                            type="button"
                            variant="ghost"
                            className="size-8 p-0 text-muted-foreground hover:text-foreground"
                        >
                            <MoreHorizontal className="size-4" />
                        </Button>
                    }
                />

                <DropdownMenuContent align="end" className="w-48 text-xs">
                    <DropdownMenuItem
                        onClick={() => regenerateMutation.mutate({ id: device.id })}
                        disabled={regenerateMutation.isPending}
                        className="gap-2 cursor-pointer"
                    >
                        <RefreshCw className="size-3.5" />
                        <span>Regenerate PIN</span>
                    </DropdownMenuItem>

                    <DropdownMenuItem
                        onClick={() => window.open("/devices/pair", "_blank")}
                        className="gap-2 cursor-pointer"
                    >
                        <ExternalLink className="size-3.5" />
                        <span>Open Pair Page</span>
                    </DropdownMenuItem>

                    <DropdownMenuItem
                        onClick={() => {
                            setName(device.name);
                            setIsUnlimitedTTL(device.expires_at == null);
                            setExpiresAt(formatForDatetimeLocal(device.expires_at));
                            setIsEditOpen(true);
                        }}
                        className="gap-2 cursor-pointer"
                    >
                        <Edit className="size-3.5" />
                        <span>Edit Device & TTL</span>
                    </DropdownMenuItem>

                    {
                        eventId ? <>
                            <DropdownMenuSeparator />
                            <DropdownMenuItem
                                onClick={() => setIsUnassignOpen(true)}
                                className="gap-2 cursor-pointer text-amber-600 dark:text-amber-400 focus:text-amber-600"
                            >
                                <Unlink className="size-3.5" />
                                <span>Remove from Event</span>
                            </DropdownMenuItem>
                        </> : null
                    }

                    <DropdownMenuSeparator />
                    <DropdownMenuItem
                        onClick={() => setIsDeleteOpen(true)}
                        className="gap-2 cursor-pointer text-destructive focus:text-destructive"
                    >
                        <Trash2 className="size-3.5" />
                        <span>Delete Device</span>
                    </DropdownMenuItem>
                </DropdownMenuContent>
            </DropdownMenu>

            {/* Edit Device Dialog */}
            <Dialog open={isEditOpen} onOpenChange={setIsEditOpen}>
                <DialogContent className="sm:max-w-md p-0 overflow-hidden flex flex-col">
                    <DialogHeader className="p-5 pb-4 border-b bg-muted/20">
                        <DialogTitle className="text-sm font-semibold">
                            Edit Scanner Device
                        </DialogTitle>
                        <DialogDescription className="text-xs text-muted-foreground mt-0.5">
                            Update terminal label and device session expiration limit.
                        </DialogDescription>
                    </DialogHeader>

                    <form onSubmit={handleSaveEdit} className="p-5 space-y-4">
                        <div className="space-y-2">
                            <Label className="text-xs font-medium">Device Name</Label>
                            <Input
                                value={name}
                                onChange={(e) => setName(e.target.value)}
                                className="text-xs h-9"
                                required
                            />
                        </div>

                        <div className="space-y-2 pt-1 border-t">
                            <Label className="text-xs font-medium flex items-center gap-1.5">
                                <Calendar className="size-3.5 text-muted-foreground" />
                                <span>Device Session Lifetime (TTL)</span>
                            </Label>

                            <div className="space-y-2">
                                <label className="flex items-center gap-2 text-xs text-muted-foreground cursor-pointer select-none">
                                    <Checkbox
                                        checked={isUnlimitedTTL}
                                        onCheckedChange={(checked) => setIsUnlimitedTTL(!!checked)}
                                    />
                                    <span>Unlimited TTL (stays verified until manually unpaired)</span>
                                </label>

                                {
                                    !isUnlimitedTTL ? <div className="space-y-1 pt-1">
                                        <Input
                                            type="datetime-local"
                                            value={expiresAt}
                                            onChange={(e) => setExpiresAt(e.target.value)}
                                            className="text-xs h-9 font-mono"
                                            required={!isUnlimitedTTL}
                                        />
                                    </div> : null
                                }
                            </div>
                        </div>

                        <DialogFooter className="pt-2 gap-2 sm:gap-0">
                            <Button
                                type="button"
                                variant="outline"
                                onClick={() => setIsEditOpen(false)}
                                className="text-xs"
                            >
                                Cancel
                            </Button>
                            <Button
                                type="submit"
                                disabled={!name.trim() || updateMutation.isPending || (!isUnlimitedTTL && !expiresAt)}
                                className="text-xs font-semibold"
                            >
                                {
                                    updateMutation.isPending ? "Saving..." : "Save Changes"
                                }
                            </Button>
                        </DialogFooter>
                    </form>
                </DialogContent>
            </Dialog>

            {/* Unassign from Event AlertDialog */}
            {
                eventId ? <AlertDialog open={isUnassignOpen} onOpenChange={setIsUnassignOpen}>
                    <AlertDialogContent>
                        <AlertDialogHeader>
                            <AlertDialogTitle className="text-sm font-semibold">
                                Remove Device from Event?
                            </AlertDialogTitle>
                            <AlertDialogDescription className="text-xs">
                                This will detach the device &quot;{device.name}&quot; from this event. The physical device record will remain preserved in your global devices pool.
                            </AlertDialogDescription>
                        </AlertDialogHeader>
                        <AlertDialogFooter>
                            <AlertDialogCancel className="text-xs">
                                Cancel
                            </AlertDialogCancel>
                            <AlertDialogAction
                                onClick={handleConfirmUnassign}
                                className="text-xs bg-amber-600 text-white hover:bg-amber-700"
                            >
                                Remove from Event
                            </AlertDialogAction>
                        </AlertDialogFooter>
                    </AlertDialogContent>
                </AlertDialog> : null
            }

            {/* Permanent Delete AlertDialog */}
            <AlertDialog open={isDeleteOpen} onOpenChange={setIsDeleteOpen}>
                <AlertDialogContent>
                    <AlertDialogHeader>
                        <AlertDialogTitle className="text-sm font-semibold text-destructive">
                            Delete Device Permanently?
                        </AlertDialogTitle>
                        <AlertDialogDescription className="text-xs">
                            This will permanently remove &quot;{device.name}&quot; and revoke its pairing session across all assigned events.
                        </AlertDialogDescription>
                    </AlertDialogHeader>
                    <AlertDialogFooter>
                        <AlertDialogCancel className="text-xs">
                            Cancel
                        </AlertDialogCancel>
                        <AlertDialogAction
                            onClick={handleConfirmDelete}
                            className="text-xs bg-destructive text-destructive-foreground hover:bg-destructive/90"
                        >
                            Yes, Delete Permanently
                        </AlertDialogAction>
                    </AlertDialogFooter>
                </AlertDialogContent>
            </AlertDialog>
        </>
    );
}
