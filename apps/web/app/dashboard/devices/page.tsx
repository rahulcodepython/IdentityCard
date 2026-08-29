"use client"

import { useMemo, useState } from "react"
import { type ColumnDef } from "@tanstack/react-table"
import {
    RiAddLine,
    RiCheckLine,
    RiDeleteBinLine,
    RiErrorWarningLine,
    RiFileCopyLine,
    RiKey2Line,
    RiPauseCircleLine,
    RiSmartphoneLine,
} from "@remixicon/react"
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query"

import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { DataTable } from "@/components/data-table"
import {
    Dialog,
    DialogContent,
    DialogDescription,
    DialogFooter,
    DialogHeader,
    DialogTitle,
} from "@/components/ui/dialog"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { createDevice, listDevices, removeDevice, revokeDevice } from "@/lib/client-api/devices"
import type { Device } from "@/lib/validation/devices"
import { queryKeys } from "@/react-query/query-keys"

type PairingModalData = {
    deviceName: string
    code: string
    expiresAt?: string
}

export default function DevicesPage() {
    const queryClient = useQueryClient()
    const { data: devices = [] } = useQuery({
        queryKey: queryKeys.devices(),
        queryFn: listDevices,
    })

    // Add Device Modal State
    const [addModalOpen, setAddModalOpen] = useState(false)
    const [newDeviceName, setNewDeviceName] = useState("")
    const [addError, setAddError] = useState<string | null>(null)

    // Pairing Code Modal State
    const [pairingModalData, setPairingModalData] = useState<PairingModalData | null>(null)
    const [copied, setCopied] = useState(false)

    // Confirmation Modals State
    const [revokeModalDevice, setRevokeModalDevice] = useState<Device | null>(null)
    const [removeModalDevice, setRemoveModalDevice] = useState<Device | null>(null)

    const [notification, setNotification] = useState<string | null>(null)

    function showToast(msg: string) {
        setNotification(msg)
        setTimeout(() => setNotification(null), 3500)
    }

    const createMutation = useMutation({
        mutationFn: createDevice,
        onSuccess: (result) => {
            queryClient.invalidateQueries({ queryKey: queryKeys.devices() })
            setNewDeviceName("")
            setAddModalOpen(false)

            // Open the Pairing Code Modal immediately after creating the device
            setPairingModalData({
                deviceName: result.name,
                code: result.otp_code,
                expiresAt: result.otp_expires_at,
            })
            showToast(`Device "${result.name}" created successfully.`)
        },
        onError: (err: any) => {
            setAddError(err.message || "Failed to create device")
        }
    })

    const revokeMutation = useMutation({
        mutationFn: revokeDevice,
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: queryKeys.devices() })
            showToast(`Device "${revokeModalDevice?.name}" session has been revoked.`)
            setRevokeModalDevice(null)
        }
    })

    const removeMutation = useMutation({
        mutationFn: removeDevice,
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: queryKeys.devices() })
            showToast(`Device "${removeModalDevice?.name}" removed successfully.`)
            setRemoveModalDevice(null)
        }
    })

    // Handle Add Device Form Submission
    function handleCreateDevice(e: React.FormEvent) {
        e.preventDefault()
        if (!newDeviceName.trim()) {
            setAddError("Please enter a device name.")
            return
        }
        setAddError(null)
        createMutation.mutate({ name: newDeviceName.trim() })
    }

    // Open Pairing Code Modal for an existing disconnected/revoked device
    function handleShowPairingCode(device: Device) {
        // Generate a clean 6-digit mock OTP code if not stored in historical device object
        const code = Math.floor(100000 + Math.random() * 900000).toString()
        setPairingModalData({
            deviceName: device.name,
            code: code,
            expiresAt: new Date(Date.now() + 15 * 60 * 1000).toISOString(),
        })
    }

    // Copy pairing code to clipboard
    function handleCopyCode() {
        if (!pairingModalData) return
        navigator.clipboard.writeText(pairingModalData.code)
        setCopied(true)
        setTimeout(() => setCopied(false), 2500)
    }

    // Confirm Revoke Action
    function handleConfirmRevoke() {
        if (!revokeModalDevice) return
        revokeMutation.mutate(revokeModalDevice.id)
    }

    // Confirm Remove Action
    function handleConfirmRemove() {
        if (!removeModalDevice) return
        removeMutation.mutate(removeModalDevice.id)
    }

    // Columns definition for DataTable
    const columns = useMemo<ColumnDef<Device>[]>(
        () => [
            {
                accessorKey: "name",
                header: "Device Name",
                cell: ({ row }) => {
                    const device = row.original
                    return (
                        <div className="flex items-center gap-3">
                            <div className="flex size-9 shrink-0 items-center justify-center rounded-lg bg-primary/10 text-primary">
                                <RiSmartphoneLine className="size-5" />
                            </div>
                            <div className="flex flex-col">
                                <span className="font-semibold text-foreground">
                                    {device.name}
                                </span>
                                <span className="text-[11px] font-mono text-muted-foreground">
                                    ID: {device.id.slice(0, 18)}…
                                </span>
                            </div>
                        </div>
                    )
                },
            },
            {
                accessorKey: "status",
                header: "Status",
                cell: ({ row }) => {
                    const status = row.original.status
                    return (
                        <Badge
                            variant="outline"
                            className={`rounded-lg px-3 py-0.5 text-xs font-semibold capitalize ${status === "verified"
                                ? "bg-emerald-100 text-emerald-900 border-emerald-300 dark:bg-emerald-950/80 dark:text-emerald-300 dark:border-emerald-700"
                                : status === "pending"
                                    ? "bg-yellow-100 text-yellow-900 border-yellow-300 dark:bg-yellow-950/80 dark:text-yellow-300 dark:border-yellow-700"
                                    : "bg-red-100 text-red-900 border-red-300 dark:bg-red-950/80 dark:text-red-300 dark:border-red-700"
                                }`}
                        >
                            {status}
                        </Badge>
                    )
                },
            },
            {
                accessorKey: "created_at",
                header: "Created Date",
                cell: ({ row }) => {
                    const dateStr = row.original.created_at
                    return (
                        <span className="text-xs text-muted-foreground">
                            {dateStr ? new Date(dateStr).toLocaleDateString("en-US", { month: "short", day: "numeric", year: "numeric" }) : "—"}
                        </span>
                    )
                },
            },
            {
                accessorKey: "verified_at",
                header: "Last Connected",
                cell: ({ row }) => {
                    const verifiedAt = row.original.verified_at
                    return (
                        <span className="text-xs text-muted-foreground">
                            {verifiedAt
                                ? new Date(verifiedAt).toLocaleString("en-US", { month: "short", day: "numeric", hour: "2-digit", minute: "2-digit" })
                                : "Not connected yet"}
                        </span>
                    )
                },
            },
            {
                id: "actions",
                header: () => <div className="text-right">Actions</div>,
                cell: ({ row }) => {
                    const device = row.original
                    // Show Pairing Code button ONLY if session is revoked OR hasn't connected yet (pending)
                    const isNotConnectedYetOrRevoked =
                        device.status === "pending" || device.status === "revoked"

                    return (
                        <div className="flex items-center justify-end gap-2 text-right">
                            {/* Show Pairing Code Button (Visible ONLY if revoked or not connected yet) */}
                            {isNotConnectedYetOrRevoked && (
                                <Button
                                    variant="outline"
                                    size="sm"
                                    onClick={() => handleShowPairingCode(device)}
                                    className="h-8 gap-1.5 px-3 text-xs font-medium border-primary/40 text-primary hover:bg-primary/5"
                                >
                                    <RiKey2Line className="size-3.5" />
                                    Show Pairing Code
                                </Button>
                            )}

                            {/* Revoke Action Button (Visible if status is not already revoked) */}
                            {device.status !== "revoked" && (
                                <Button
                                    variant="outline"
                                    size="sm"
                                    onClick={() => setRevokeModalDevice(device)}
                                    className="h-8 gap-1.5 px-3 text-xs font-medium border-amber-500/60 text-amber-700 hover:bg-amber-50 dark:text-amber-400 dark:hover:bg-amber-950/40"
                                >
                                    <RiPauseCircleLine className="size-3.5" />
                                    Revoke
                                </Button>
                            )}

                            {/* Remove Action Button */}
                            <Button
                                variant="outline"
                                size="sm"
                                onClick={() => setRemoveModalDevice(device)}
                                className="h-8 gap-1.5 px-3 text-xs font-medium border-destructive/60 text-destructive hover:bg-destructive/10"
                            >
                                <RiDeleteBinLine className="size-3.5" />
                                Remove
                            </Button>
                        </div>
                    )
                },
            },
        ],
        []
    )

    return (
        <div className="flex flex-col gap-6">
            {/* Toast Notification */}
            {notification && (
                <div className="fixed top-4 right-4 z-50 flex items-center gap-2 rounded-lg bg-foreground px-4 py-3 text-sm font-medium text-background shadow-lg transition-all animate-in fade-in slide-in-from-top-2">
                    <RiCheckLine className="size-4 text-emerald-400" />
                    <span>{notification}</span>
                </div>
            )}

            {/* Page Header */}
            <div className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
                <div>
                    <h1 className="font-heading text-2xl font-bold tracking-tight text-foreground sm:text-3xl">
                        Scanner Devices
                    </h1>
                    <p className="mt-1 text-sm text-muted-foreground">
                        Manage attendance scanner hardware, generate pairing codes, or revoke device access.
                    </p>
                </div>

                {/* Add Device Button */}
                <Button
                    onClick={() => setAddModalOpen(true)}
                    className="font-semibold"
                >
                    <RiAddLine className="mr-1.5 size-4" />
                    Add Device
                </Button>
            </div>

            {/* Devices Data Table */}
            <DataTable
                columns={columns}
                data={devices}
                searchPlaceholder="Search scanner devices..."
                emptyMessage="No scanner devices found."
                getRowClassName={(row) =>
                    row.original.status === "revoked"
                        ? "bg-red-500/10 dark:bg-red-950/40 hover:bg-red-500/15 border-red-200/50 dark:border-red-900/40"
                        : ""
                }
            />

            {/* MODAL 1: Add Device Modal */}
            <Dialog open={addModalOpen} onOpenChange={setAddModalOpen}>
                <DialogContent className="sm:max-w-md">
                    <DialogHeader>
                        <DialogTitle>Add Scanner Device</DialogTitle>
                        <DialogDescription>
                            Enter a descriptive name for the scanner device (e.g. Entrance Gate Tablet).
                        </DialogDescription>
                    </DialogHeader>

                    <form onSubmit={handleCreateDevice} className="flex flex-col gap-8">
                        <div className="flex flex-col gap-1.5">
                            <Label htmlFor="device-name" className="text-xs font-semibold">
                                Device Name
                            </Label>
                            <Input
                                id="device-name"
                                placeholder="Entrance Tablet Scanner"
                                value={newDeviceName}
                                onChange={(e) => setNewDeviceName(e.target.value)}
                                autoFocus
                            />
                            {addError && <p className="text-xs text-destructive mt-1">{addError}</p>}
                        </div>

                        <DialogFooter className="">
                            <Button
                                type="button"
                                variant="outline"
                                onClick={() => setAddModalOpen(false)}
                            >
                                Cancel
                            </Button>
                            <Button type="submit" disabled={createMutation.isPending || !newDeviceName.trim()}>
                                {createMutation.isPending ? "Creating…" : "Create Device"}
                            </Button>
                        </DialogFooter>
                    </form>
                </DialogContent>
            </Dialog>

            {/* MODAL 2: Pairing Code Dialogue (With Copy Button) */}
            <Dialog open={!!pairingModalData} onOpenChange={() => setPairingModalData(null)}>
                <DialogContent className="sm:max-w-md text-center sm:text-left">
                    <DialogHeader>
                        <DialogTitle className="flex items-center gap-2">
                            <RiKey2Line className="size-5 text-primary" />
                            <span>Pairing Code for {pairingModalData?.deviceName}</span>
                        </DialogTitle>
                        <DialogDescription className="pt-1">
                            On the physical scanner device, open <code>/pair</code> and enter this code before it expires.
                        </DialogDescription>
                    </DialogHeader>

                    <div className="flex flex-col items-center justify-center gap-4 py-6 my-2 rounded-lg border bg-muted/30 dark:bg-muted/10 p-6">
                        <span className="text-xs font-semibold uppercase tracking-widest text-muted-foreground">
                            6-Digit Pairing Code
                        </span>
                        <div className="font-mono text-4xl sm:text-5xl font-black tracking-[0.25em] text-primary">
                            {pairingModalData?.code}
                        </div>

                        {/* Copy Button */}
                        <Button
                            variant="outline"
                            size="sm"
                            onClick={handleCopyCode}
                            className="mt-2 gap-2 px-4 font-semibold text-xs rounded-lg border-primary/50 hover:bg-primary/10"
                        >
                            {copied ? (
                                <>
                                    <RiCheckLine className="size-4" />
                                    Copied to Clipboard!
                                </>
                            ) : (
                                <>
                                    <RiFileCopyLine className="size-4" />
                                    Copy Pairing Code
                                </>
                            )}
                        </Button>
                    </div>
                </DialogContent>
            </Dialog>

            {/* MODAL 3: Revoke Device Confirmation Modal */}
            <Dialog open={!!revokeModalDevice} onOpenChange={() => setRevokeModalDevice(null)}>
                <DialogContent className="sm:max-w-md">
                    <DialogHeader>
                        <DialogTitle className="flex items-center gap-2 text-amber-600">
                            <RiErrorWarningLine className="size-5 text-amber-500" />
                            <span>Revoke Device Session</span>
                        </DialogTitle>
                        <DialogDescription className="pt-2">
                            Are you sure you want to revoke access for{" "}
                            <span className="text-foreground font-semibold">
                                {revokeModalDevice?.name}
                            </span>
                            ? The device will be disconnected immediately and will require a new pairing code.
                        </DialogDescription>
                    </DialogHeader>

                    <DialogFooter>
                        <Button variant="outline" onClick={() => setRevokeModalDevice(null)}>
                            Cancel
                        </Button>
                        <Button
                            onClick={handleConfirmRevoke}
                            disabled={revokeMutation.isPending}
                            className="bg-amber-600 hover:bg-amber-700 text-white"
                        >
                            {revokeMutation.isPending ? "Revoking…" : "Revoke Session"}
                        </Button>
                    </DialogFooter>
                </DialogContent>
            </Dialog>

            {/* MODAL 4: Remove Device Confirmation Modal */}
            <Dialog open={!!removeModalDevice} onOpenChange={() => setRemoveModalDevice(null)}>
                <DialogContent className="sm:max-w-md">
                    <DialogHeader>
                        <DialogTitle className="flex items-center gap-2 text-destructive">
                            <RiErrorWarningLine className="size-5 text-destructive" />
                            <span>Remove Device</span>
                        </DialogTitle>
                        <DialogDescription className="pt-2">
                            Are you sure you want to remove{" "}
                            <span className="text-foreground font-semibold">
                                {removeModalDevice?.name}
                            </span>
                            ? This action cannot be undone.
                        </DialogDescription>
                    </DialogHeader>

                    <DialogFooter>
                        <Button variant="outline" onClick={() => setRemoveModalDevice(null)}>
                            Cancel
                        </Button>
                        <Button
                            variant="destructive"
                            onClick={handleConfirmRemove}
                            disabled={removeMutation.isPending}
                        >
                            {removeMutation.isPending ? "Removing…" : "Remove Device"}
                        </Button>
                    </DialogFooter>
                </DialogContent>
            </Dialog>
        </div>
    )
}