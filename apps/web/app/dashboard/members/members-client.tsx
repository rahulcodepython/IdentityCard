"use client"

import { useMemo, useState } from "react"
import { type ColumnDef } from "@tanstack/react-table"
import {
    RiCalendarEventLine,
    RiCheckLine,
    RiDeleteBinLine,
    RiDownloadLine,
    RiErrorWarningLine,
    RiFilter3Line,
    RiMailSendLine,
    RiPauseCircleLine,
    RiPlayCircleLine,
    RiUserAddLine,
} from "@remixicon/react"

import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Checkbox } from "@/components/ui/checkbox"
import { DataTable } from "@/components/ui/data-table"
import {
    Dialog,
    DialogContent,
    DialogDescription,
    DialogFooter,
    DialogHeader,
    DialogTitle,
} from "@/components/ui/dialog"
import {
    DropdownMenu,
    DropdownMenuCheckboxItem,
    DropdownMenuContent,
    DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import {
    Select,
    SelectContent,
    SelectItem,
    SelectTrigger,
    SelectValue,
} from "@/components/ui/select"

type Status = "Active" | "Suspended"

type Member = {
    id: string
    name: string
    email: string
    enrolledDate: string
    status: Status
    assignedEvents: string[]
}

type EventItem = {
    id: string
    name: string
    date: string
}

const ALL_EVENTS: EventItem[] = [
    { id: "e1", name: "TechConf 2026", date: "Mar 15, 2026" },
    { id: "e2", name: "Annual Developer Summit", date: "Apr 20, 2026" },
    { id: "e3", name: "Flash Meetup 2026", date: "May 05, 2026" },
    { id: "e4", name: "University Hackathon 2026", date: "Jun 12, 2026" },
    { id: "e5", name: "Global Tech Expo", date: "Jul 18, 2026" },
]

const INITIAL_MEMBERS: Member[] = [
    {
        id: "m1",
        name: "Sarah Jenkins",
        email: "sarah.j@example.com",
        enrolledDate: "Jan 10, 2026",
        status: "Active",
        assignedEvents: ["TechConf 2026", "Annual Developer Summit"],
    },
    {
        id: "m2",
        name: "David Chen",
        email: "david.chen@example.com",
        enrolledDate: "Jan 14, 2026",
        status: "Active",
        assignedEvents: ["TechConf 2026"],
    },
    {
        id: "m3",
        name: "Elena Rostova",
        email: "elena.r@example.com",
        enrolledDate: "Feb 01, 2026",
        status: "Suspended",
        assignedEvents: ["Flash Meetup 2026"],
    },
    {
        id: "m4",
        name: "Marcus Vance",
        email: "marcus.vance@example.com",
        enrolledDate: "Feb 08, 2026",
        status: "Active",
        assignedEvents: ["University Hackathon 2026", "Global Tech Expo"],
    },
    {
        id: "m5",
        name: "Aisha Patel",
        email: "aisha.patel@example.com",
        enrolledDate: "Feb 12, 2026",
        status: "Active",
        assignedEvents: ["TechConf 2026", "Global Tech Expo"],
    },
]

export function MembersClient() {
    const [members, setMembers] = useState<Member[]>(INITIAL_MEMBERS)
    const [statusFilter, setStatusFilter] = useState<string>("All")
    const [selectedEventFilters, setSelectedEventFilters] = useState<string[]>([])

    // Modals state
    const [inviteOpen, setInviteOpen] = useState(false)
    const [inviteEmail, setInviteEmail] = useState("")
    const [inviteSelectedEvents, setInviteSelectedEvents] = useState<string[]>([])

    const [assignMember, setAssignMember] = useState<Member | null>(null)
    const [assignSelectedEvents, setAssignSelectedEvents] = useState<string[]>([])

    const [suspendMember, setSuspendMember] = useState<Member | null>(null)
    const [deleteMember, setDeleteMember] = useState<Member | null>(null)

    const [notification, setNotification] = useState<string | null>(null)

    function showToast(msg: string) {
        setNotification(msg)
        setTimeout(() => setNotification(null), 3500)
    }

    // Filtered members calculation based on status and multiple selected events
    const filteredMembers = useMemo(() => {
        return members.filter((m) => {
            const matchesStatus =
                statusFilter === "All" || m.status === statusFilter

            const matchesEvents =
                selectedEventFilters.length === 0 ||
                m.assignedEvents.some((ev) => selectedEventFilters.includes(ev))

            return matchesStatus && matchesEvents
        })
    }, [members, statusFilter, selectedEventFilters])

    // Download filtered data as CSV
    function handleDownloadCSV() {
        const headers = ["ID", "Name", "Email", "Enrolled Date", "Status", "Assigned Events"]
        const rows = filteredMembers.map((m) => [
            m.id,
            `"${m.name.replace(/"/g, '""')}"`,
            `"${m.email.replace(/"/g, '""')}"`,
            m.enrolledDate,
            m.status,
            `"${m.assignedEvents.join("; ").replace(/"/g, '""')}"`,
        ])

        const csvContent = [headers.join(","), ...rows.map((r) => r.join(","))].join("\n")
        const blob = new Blob([csvContent], { type: "text/csv;charset=utf-8;" })
        const url = URL.createObjectURL(blob)
        const link = document.createElement("a")
        link.setAttribute("href", url)
        link.setAttribute("download", `organization_members_${Date.now()}.csv`)
        document.body.appendChild(link)
        link.click()
        document.body.removeChild(link)
        showToast(`Downloaded CSV with ${filteredMembers.length} member(s)`)
    }

    // Handle Invite Member (Role removed per request)
    function handleSendInvite() {
        if (!inviteEmail.trim()) return
        const nameFromEmail = inviteEmail.split("@")[0]
        const formattedName = nameFromEmail
            .replace(/[._]/g, " ")
            .replace(/\b\w/g, (l) => l.toUpperCase())

        const newMember: Member = {
            id: `m_${Date.now()}`,
            name: formattedName || "Invited Member",
            email: inviteEmail.trim(),
            enrolledDate: new Date().toLocaleDateString("en-US", { month: "short", day: "2-digit", year: "numeric" }),
            status: "Active",
            assignedEvents: inviteSelectedEvents,
        }

        setMembers((prev) => [newMember, ...prev])
        showToast(`Invitation sent to ${inviteEmail}`)
        setInviteEmail("")
        setInviteSelectedEvents([])
        setInviteOpen(false)
    }

    // Open Assign Events Modal
    function openAssignModal(member: Member) {
        setAssignMember(member)
        setAssignSelectedEvents([...member.assignedEvents])
    }

    // Save Event Assignments
    function handleSaveAssignments() {
        if (!assignMember) return
        setMembers((prev) =>
            prev.map((m) =>
                m.id === assignMember.id
                    ? { ...m, assignedEvents: [...assignSelectedEvents] }
                    : m
            )
        )
        showToast(`Updated event assignments for ${assignMember.name}`)
        setAssignMember(null)
    }

    // Confirm Suspension
    function handleConfirmSuspend() {
        if (!suspendMember) return
        const newStatus: Status = suspendMember.status === "Active" ? "Suspended" : "Active"
        setMembers((prev) =>
            prev.map((m) => (m.id === suspendMember.id ? { ...m, status: newStatus } : m))
        )
        showToast(
            `${suspendMember.name} has been ${newStatus === "Suspended" ? "suspended" : "reactivated"}`
        )
        setSuspendMember(null)
    }

    // Confirm Delete
    function handleConfirmDelete() {
        if (!deleteMember) return
        setMembers((prev) => prev.filter((m) => m.id !== deleteMember.id))
        showToast(`${deleteMember.name} has been deleted`)
        setDeleteMember(null)
    }

    // DataTable Columns configuration (Status column removed per request)
    const columns = useMemo<ColumnDef<Member>[]>(
        () => [
            {
                accessorKey: "name",
                header: "Member",
                cell: ({ row }) => {
                    const member = row.original
                    return (
                        <div className="flex items-center gap-3">
                            <div className="flex size-9 shrink-0 items-center justify-center rounded-full bg-primary/10 font-semibold text-primary text-sm">
                                {member.name.charAt(0)}
                            </div>
                            <div className="flex flex-col">
                                <div className="flex items-center gap-2">
                                    <span className="font-semibold text-foreground">
                                        {member.name}
                                    </span>
                                    {member.status === "Suspended" && (
                                        <Badge variant="destructive" className="h-4 px-1.5 text-[10px] uppercase font-bold">
                                            Suspended
                                        </Badge>
                                    )}
                                </div>
                                <span className="text-xs text-muted-foreground">
                                    {member.email}
                                </span>
                            </div>
                        </div>
                    )
                },
            },
            {
                accessorKey: "enrolledDate",
                header: "Enrolled Date",
                cell: ({ row }) => (
                    <span className="text-xs text-muted-foreground">
                        {row.original.enrolledDate}
                    </span>
                ),
            },
            {
                accessorKey: "assignedEvents",
                header: "Assigned Events",
                cell: ({ row }) => {
                    const events = row.original.assignedEvents
                    return events.length > 0 ? (
                        <div className="flex flex-wrap gap-1">
                            {events.map((ev) => (
                                <Badge key={ev} variant="secondary" className="text-[11px] font-normal">
                                    {ev}
                                </Badge>
                            ))}
                        </div>
                    ) : (
                        <span className="text-xs text-muted-foreground italic">
                            No events assigned
                        </span>
                    )
                },
            },
            {
                id: "actions",
                header: () => <div className="text-right">Actions</div>,
                cell: ({ row }) => {
                    const member = row.original
                    return (
                        <div className="flex items-center justify-end gap-2 text-right">
                            {/* Action 1: Assign Events (Outline) */}
                            <Button
                                variant="outline"
                                size="sm"
                                onClick={() => openAssignModal(member)}
                                className="h-8 gap-1.5 px-3 text-xs font-medium"
                            >
                                <RiCalendarEventLine className="size-3.5 text-muted-foreground" />
                                Assign Events
                            </Button>

                            {/* Action 2: Suspend / Reactivate (Outline) */}
                            <Button
                                variant="outline"
                                size="sm"
                                onClick={() => setSuspendMember(member)}
                                className={`h-8 gap-1.5 px-3 text-xs font-medium ${member.status === "Active"
                                    ? "border-amber-500/60 text-amber-700 hover:bg-amber-50 dark:text-amber-400 dark:hover:bg-amber-950/40"
                                    : "border-emerald-500/60 text-emerald-700 hover:bg-emerald-50 dark:text-emerald-400 dark:hover:bg-emerald-950/40"
                                    }`}
                            >
                                {member.status === "Active" ? (
                                    <>
                                        <RiPauseCircleLine className="size-3.5" />
                                        Suspend
                                    </>
                                ) : (
                                    <>
                                        <RiPlayCircleLine className="size-3.5" />
                                        Reactivate
                                    </>
                                )}
                            </Button>

                            {/* Action 3: Delete (Outline) */}
                            <Button
                                variant="outline"
                                size="sm"
                                onClick={() => setDeleteMember(member)}
                                className="h-8 gap-1.5 px-3 text-xs font-medium border-destructive/60 text-destructive hover:bg-destructive/10"
                            >
                                <RiDeleteBinLine className="size-3.5" />
                                Delete
                            </Button>
                        </div>
                    )
                },
            },
        ],
        []
    )

    // Extra toolbar controls passed directly beside Column Viewer in DataTable
    const extraToolbarControls = (
        <>
            {/* Multi-Select Events Filter (Wider content container so event names fit in 1 line) */}
            <DropdownMenu>
                <DropdownMenuTrigger
                    render={
                        <Button variant="outline" size="sm">
                            <RiFilter3Line className="size-3.5 text-muted-foreground" />
                            Events ({selectedEventFilters.length ? selectedEventFilters.length : "All"})
                        </Button>
                    }
                />
                <DropdownMenuContent align="end" className="w-80 max-w-sm whitespace-nowrap p-1.5">
                    {ALL_EVENTS.map((ev) => {
                        const checked = selectedEventFilters.includes(ev.name)
                        return (
                            <DropdownMenuCheckboxItem
                                key={ev.id}
                                checked={checked}
                                onCheckedChange={(c) => {
                                    if (c) {
                                        setSelectedEventFilters((prev) => [...prev, ev.name])
                                    } else {
                                        setSelectedEventFilters((prev) =>
                                            prev.filter((n) => n !== ev.name)
                                        )
                                    }
                                }}
                                onSelect={(e) => e.preventDefault()}
                                className="cursor-pointer whitespace-nowrap text-xs flex items-center justify-between py-2 px-2.5"
                            >
                                <span>{ev.name}</span>
                                <span className="ml-3 text-[11px] text-muted-foreground font-normal">
                                    {ev.date}
                                </span>
                            </DropdownMenuCheckboxItem>
                        )
                    })}
                </DropdownMenuContent>
            </DropdownMenu>

            {/* Status Select Filter using Shadcn UI Select */}
            <Select value={statusFilter} onValueChange={(val) => setStatusFilter(val ?? "All")}>
                <SelectTrigger size="sm" className="h-8 w-32 text-xs">
                    <SelectValue placeholder="All Statuses" />
                </SelectTrigger>
                <SelectContent>
                    <SelectItem value="All">All Statuses</SelectItem>
                    <SelectItem value="Active">Active</SelectItem>
                    <SelectItem value="Suspended">Suspended</SelectItem>
                </SelectContent>
            </Select>

            {/* Download Data Button */}
            <Button
                variant="outline"
                size="sm"
                onClick={handleDownloadCSV}
            >
                <RiDownloadLine className="size-3.5 text-muted-foreground" />
                Download Data
            </Button>
        </>
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
                        Members
                    </h1>
                    <p className="mt-1 text-sm text-muted-foreground">
                        Manage organization members, invite new colleagues, assign selective event access, or manage accounts.
                    </p>
                </div>

                {/* Invite Member Button */}
                <Button
                    onClick={() => setInviteOpen(true)}
                    className="font-semibold"
                >
                    <RiUserAddLine className="mr-1.5 size-4" />
                    Invite Member
                </Button>
            </div>

            {/* Members Data Table Component with Toolbar Filters & Suspended Row Red Highlight */}
            <DataTable
                columns={columns}
                data={filteredMembers}
                searchPlaceholder="Search member name or email..."
                emptyMessage="No members found matching your search or filters."
                extraActions={extraToolbarControls}
                getRowClassName={(row) =>
                    row.original.status === "Suspended"
                        ? "bg-red-500/10 dark:bg-red-950/40 hover:bg-red-500/15 border-red-200/50 dark:border-red-900/40"
                        : ""
                }
            />

            {/* MODAL 1: Invite Member Modal */}
            <Dialog open={inviteOpen} onOpenChange={setInviteOpen}>
                <DialogContent className="sm:max-w-lg min-h-[540px] max-h-[85vh] flex flex-col justify-between overflow-y-auto">
                    <DialogHeader>
                        <DialogTitle>Invite New Member</DialogTitle>
                        <DialogDescription>
                            Enter the email address of the colleague you want to invite to your organization and select initial event access.
                        </DialogDescription>
                    </DialogHeader>

                    <div className="flex flex-col gap-4 py-3 flex-1">
                        <div className="flex flex-col gap-1.5">
                            <Label htmlFor="invite-email" className="font-semibold text-xs text-foreground">
                                Email Address
                            </Label>
                            <Input
                                id="invite-email"
                                type="email"
                                placeholder="colleague@example.com"
                                value={inviteEmail}
                                onChange={(e) => setInviteEmail(e.target.value)}
                            />
                        </div>

                        {/* Selective Events Assignment checkboxes in invite modal */}
                        <div className="flex flex-col gap-2 border-t pt-3 flex-1">
                            <Label className="text-xs font-semibold text-muted-foreground uppercase tracking-wider">
                                Assign Initial Events (Optional)
                            </Label>
                            <div className="flex flex-col gap-2 max-h-48 overflow-y-auto pr-1 border rounded-xl p-2.5 bg-muted/20">
                                {ALL_EVENTS.map((ev) => {
                                    const checked = inviteSelectedEvents.includes(ev.name)
                                    return (
                                        <label
                                            key={ev.id}
                                            className={`flex items-center gap-3 rounded-lg border p-2.5 text-xs font-medium cursor-pointer transition-all ${checked
                                                ? "border-primary/80 bg-primary/5 shadow-2xs"
                                                : "border-border/60 hover:bg-muted/40"
                                                }`}
                                        >
                                            <Checkbox
                                                checked={checked}
                                                onCheckedChange={(c) => {
                                                    if (c) {
                                                        setInviteSelectedEvents((prev) => [...prev, ev.name])
                                                    } else {
                                                        setInviteSelectedEvents((prev) =>
                                                            prev.filter((name) => name !== ev.name)
                                                        )
                                                    }
                                                }}
                                            />
                                            <div className="flex flex-col">
                                                <span className="text-foreground">{ev.name}</span>
                                                <span className="text-[11px] text-muted-foreground font-normal">
                                                    Scheduled for {ev.date}
                                                </span>
                                            </div>
                                        </label>
                                    )
                                })}
                            </div>
                        </div>
                    </div>

                    <DialogFooter>
                        <Button variant="outline" onClick={() => setInviteOpen(false)}>
                            Cancel
                        </Button>
                        <Button onClick={handleSendInvite} disabled={!inviteEmail.trim()}>
                            <RiMailSendLine className="mr-1.5 size-4" />
                            Send Invite
                        </Button>
                    </DialogFooter>
                </DialogContent>
            </Dialog>

            {/* MODAL 2: Assign Events Modal */}
            <Dialog open={!!assignMember} onOpenChange={() => setAssignMember(null)}>
                <DialogContent className="sm:max-w-md min-h-[380px] max-h-[85vh] flex flex-col justify-between overflow-y-auto">
                    <DialogHeader>
                        <DialogTitle>Assign Events to {assignMember?.name}</DialogTitle>
                        <DialogDescription>
                            Selectively check the events this member should be allowed to access or manage.
                        </DialogDescription>
                    </DialogHeader>

                    <div className="flex flex-col gap-2.5 py-2 flex-1">
                        <Label className="text-xs font-semibold text-muted-foreground uppercase tracking-wider">
                            Available Events List
                        </Label>

                        <div className="flex flex-col gap-2 max-h-64 overflow-y-auto pr-1">
                            {ALL_EVENTS.map((ev) => {
                                const isChecked = assignSelectedEvents.includes(ev.name)
                                return (
                                    <label
                                        key={ev.id}
                                        className={`flex items-center gap-3 rounded-xl border p-3 text-sm font-medium cursor-pointer transition-all ${isChecked
                                            ? "border-primary/80 bg-primary/5 shadow-2xs"
                                            : "border-border/70 hover:bg-muted/40"
                                            }`}
                                    >
                                        <Checkbox
                                            checked={isChecked}
                                            onCheckedChange={(checked) => {
                                                if (checked) {
                                                    setAssignSelectedEvents((prev) => [...prev, ev.name])
                                                } else {
                                                    setAssignSelectedEvents((prev) =>
                                                        prev.filter((n) => n !== ev.name)
                                                    )
                                                }
                                            }}
                                        />
                                        <div className="flex flex-col">
                                            <span className="text-foreground">{ev.name}</span>
                                            <span className="text-xs text-muted-foreground font-normal">
                                                Scheduled for {ev.date}
                                            </span>
                                        </div>
                                    </label>
                                )
                            })}
                        </div>
                    </div>

                    <DialogFooter>
                        <Button variant="outline" onClick={() => setAssignMember(null)}>
                            Cancel
                        </Button>
                        <Button onClick={handleSaveAssignments}>
                            Save Assignments
                        </Button>
                    </DialogFooter>
                </DialogContent>
            </Dialog>

            {/* MODAL 3: Suspend Confirmation Modal */}
            <Dialog open={!!suspendMember} onOpenChange={() => setSuspendMember(null)}>
                <DialogContent className="sm:max-w-md">
                    <DialogHeader>
                        <DialogTitle className="flex items-center gap-2">
                            <RiErrorWarningLine className="size-5 text-amber-500" />
                            <span>
                                {suspendMember?.status === "Active"
                                    ? "Confirm Suspension"
                                    : "Reactivate Member"}
                            </span>
                        </DialogTitle>
                        <DialogDescription className="pt-2">
                            Are you sure you want to{" "}
                            <strong>
                                {suspendMember?.status === "Active" ? "suspend" : "reactivate"}
                            </strong>{" "}
                            <span className="text-foreground font-semibold">{suspendMember?.name}</span>?
                            {suspendMember?.status === "Active" &&
                                " Suspended members will temporarily lose access to all organization events."}
                        </DialogDescription>
                    </DialogHeader>

                    <DialogFooter>
                        <Button variant="outline" onClick={() => setSuspendMember(null)}>
                            Cancel
                        </Button>
                        <Button
                            onClick={handleConfirmSuspend}
                            className={
                                suspendMember?.status === "Active"
                                    ? "bg-amber-600 hover:bg-amber-700 text-white"
                                    : "bg-emerald-600 hover:bg-emerald-700 text-white"
                            }
                        >
                            {suspendMember?.status === "Active"
                                ? "Suspend Member"
                                : "Reactivate Member"}
                        </Button>
                    </DialogFooter>
                </DialogContent>
            </Dialog>

            {/* MODAL 4: Delete Confirmation Modal */}
            <Dialog open={!!deleteMember} onOpenChange={() => setDeleteMember(null)}>
                <DialogContent className="sm:max-w-md">
                    <DialogHeader>
                        <DialogTitle className="flex items-center gap-2 text-destructive">
                            <RiErrorWarningLine className="size-5 text-destructive" />
                            <span>Delete Member</span>
                        </DialogTitle>
                        <DialogDescription className="pt-2">
                            Are you sure you want to delete{" "}
                            <span className="text-foreground font-semibold">{deleteMember?.name}</span>?
                            This action cannot be undone and will permanently remove their access.
                        </DialogDescription>
                    </DialogHeader>

                    <DialogFooter>
                        <Button variant="outline" onClick={() => setDeleteMember(null)}>
                            Cancel
                        </Button>
                        <Button variant="destructive" onClick={handleConfirmDelete}>
                            Delete Member
                        </Button>
                    </DialogFooter>
                </DialogContent>
            </Dialog>
        </div>
    )
}
