"use client"

import { useState } from "react"
import { useQuery, useQueryClient } from "@tanstack/react-query"
import { toast } from "sonner"
import { RiCloseLine, RiDeleteBinLine, RiMailSendLine, RiUserAddLine } from "@remixicon/react"

import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
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
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select"
import { authClient } from "@/lib/auth-client"
import { useSessionStore } from "@/store/session.store"

// A member holds exactly one role — super_admin is reserved for the org
// creator (see lib/auth.ts's creatorRole) and never offered here, so
// this page can't accidentally create a second one. Real invite/list/
// role-change/remove calls against better-auth's organization plugin —
// nothing mocked, unlike the version this replaced.
const INVITE_ROLES = [
    { value: "admin", label: "Admin" },
    { value: "scanner", label: "Scanner" },
] as const

export default function MembersPage() {
    const queryClient = useQueryClient()
    const currentUserId = useSessionStore((s) => s.user?.id)

    const membersQuery = useQuery({
        queryKey: ["org-members"],
        queryFn: async () => {
            const { data, error } = await authClient.organization.listMembers({ query: { limit: 100 } })
            if (error) throw new Error(error.message)
            return data
        },
    })

    const invitationsQuery = useQuery({
        queryKey: ["org-invitations"],
        queryFn: async () => {
            const { data, error } = await authClient.organization.listInvitations()
            if (error) throw new Error(error.message)
            return data
        },
    })

    const [inviteOpen, setInviteOpen] = useState(false)
    const [inviteEmail, setInviteEmail] = useState("")
    const [inviteRole, setInviteRole] = useState<(typeof INVITE_ROLES)[number]["value"]>("admin")
    const [invitePending, setInvitePending] = useState(false)
    const [removeTarget, setRemoveTarget] = useState<{ id: string; email: string } | null>(null)

    function invalidate() {
        void queryClient.invalidateQueries({ queryKey: ["org-members"] })
        void queryClient.invalidateQueries({ queryKey: ["org-invitations"] })
    }

    async function sendInvite() {
        if (!inviteEmail.trim()) return
        setInvitePending(true)
        const { error } = await authClient.organization.inviteMember({
            email: inviteEmail.trim(),
            role: inviteRole,
        })
        setInvitePending(false)
        if (error) {
            toast.error(error.message ?? "Couldn't send invitation.")
            return
        }
        toast.success(`Invitation sent to ${inviteEmail}`)
        setInviteEmail("")
        setInviteOpen(false)
        invalidate()
    }

    async function cancelInvitation(id: string) {
        const { error } = await authClient.organization.cancelInvitation({ invitationId: id })
        if (error) {
            toast.error(error.message ?? "Couldn't cancel invitation.")
            return
        }
        toast.success("Invitation canceled")
        invalidate()
    }

    async function changeRole(memberId: string, role: string) {
        const { error } = await authClient.organization.updateMemberRole({ memberId, role })
        if (error) {
            toast.error(error.message ?? "Couldn't update role.")
            return
        }
        toast.success("Role updated")
        invalidate()
    }

    async function confirmRemove() {
        if (!removeTarget) return
        const { error } = await authClient.organization.removeMember({ memberIdOrEmail: removeTarget.id })
        if (error) {
            toast.error(error.message ?? "Couldn't remove member.")
        } else {
            toast.success(`${removeTarget.email} removed`)
        }
        setRemoveTarget(null)
        invalidate()
    }

    const members = membersQuery.data?.members ?? []
    const invitations = (invitationsQuery.data ?? []).filter((inv) => inv.status === "pending")

    return (
        <div className="flex flex-col gap-6">
            <div className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
                <div>
                    <h1 className="font-heading text-2xl font-bold tracking-tight text-foreground sm:text-3xl">
                        Members
                    </h1>
                    <p className="mt-1 text-sm text-muted-foreground">
                        Manage who has access to this organization and what they can do.
                    </p>
                </div>
                <Button onClick={() => setInviteOpen(true)} className="font-semibold">
                    <RiUserAddLine className="mr-1.5 size-4" />
                    Invite Member
                </Button>
            </div>

            <div className="overflow-hidden rounded-lg border">
                <table className="w-full text-sm">
                    <thead className="border-b bg-muted/40 text-xs uppercase text-muted-foreground">
                        <tr>
                            <th className="px-4 py-2.5 text-left font-medium">Member</th>
                            <th className="px-4 py-2.5 text-left font-medium">Role</th>
                            <th className="px-4 py-2.5 text-right font-medium">Actions</th>
                        </tr>
                    </thead>
                    <tbody>
                        {members.length === 0 && (
                            <tr>
                                <td colSpan={3} className="px-4 py-8 text-center text-sm text-muted-foreground">
                                    {membersQuery.isLoading ? "Loading…" : "No members found."}
                                </td>
                            </tr>
                        )}
                        {members.map((member) => {
                            const isSelf = member.userId === currentUserId
                            const isOwner = member.role === "super_admin"
                            return (
                                <tr key={member.id} className="border-b last:border-0">
                                    <td className="px-4 py-3">
                                        <div className="flex flex-col">
                                            <span className="font-medium text-foreground">{member.user.name}</span>
                                            <span className="text-xs text-muted-foreground">{member.user.email}</span>
                                        </div>
                                    </td>
                                    <td className="px-4 py-3">
                                        {isOwner ? (
                                            <Badge variant="secondary">Super Admin</Badge>
                                        ) : (
                                            <Select
                                                value={member.role}
                                                onValueChange={(role) => role && void changeRole(member.id, role)}
                                            >
                                                <SelectTrigger size="sm" className="h-8 w-32 text-xs capitalize">
                                                    <SelectValue />
                                                </SelectTrigger>
                                                <SelectContent>
                                                    {INVITE_ROLES.map((r) => (
                                                        <SelectItem key={r.value} value={r.value}>
                                                            {r.label}
                                                        </SelectItem>
                                                    ))}
                                                </SelectContent>
                                            </Select>
                                        )}
                                    </td>
                                    <td className="px-4 py-3 text-right">
                                        {!isOwner && !isSelf && (
                                            <Button
                                                variant="outline"
                                                size="sm"
                                                onClick={() => setRemoveTarget({ id: member.id, email: member.user.email })}
                                                className="h-8 gap-1.5 border-destructive/60 text-destructive hover:bg-destructive/10"
                                            >
                                                <RiDeleteBinLine className="size-3.5" />
                                                Remove
                                            </Button>
                                        )}
                                    </td>
                                </tr>
                            )
                        })}
                    </tbody>
                </table>
            </div>

            {invitations.length > 0 && (
                <div className="flex flex-col gap-2">
                    <h2 className="text-sm font-semibold text-foreground">Pending invitations</h2>
                    <div className="divide-y rounded-lg border">
                        {invitations.map((inv) => (
                            <div key={inv.id} className="flex items-center justify-between px-4 py-3 text-sm">
                                <div className="flex flex-col">
                                    <span className="font-medium text-foreground">{inv.email}</span>
                                    <span className="text-xs capitalize text-muted-foreground">{inv.role}</span>
                                </div>
                                <Button
                                    variant="ghost"
                                    size="sm"
                                    onClick={() => void cancelInvitation(inv.id)}
                                    className="h-8 gap-1.5 text-muted-foreground"
                                >
                                    <RiCloseLine className="size-3.5" />
                                    Cancel
                                </Button>
                            </div>
                        ))}
                    </div>
                </div>
            )}

            <Dialog open={inviteOpen} onOpenChange={setInviteOpen}>
                <DialogContent className="sm:max-w-md">
                    <DialogHeader>
                        <DialogTitle>Invite a member</DialogTitle>
                        <DialogDescription>They&apos;ll get an email with a link to join.</DialogDescription>
                    </DialogHeader>
                    <div className="flex flex-col gap-4">
                        <div className="flex flex-col gap-1.5">
                            <Label htmlFor="invite-email">Email</Label>
                            <Input
                                id="invite-email"
                                type="email"
                                value={inviteEmail}
                                onChange={(e) => setInviteEmail(e.target.value)}
                                placeholder="colleague@example.com"
                            />
                        </div>
                        <div className="flex flex-col gap-1.5">
                            <Label>Role</Label>
                            <Select
                                value={inviteRole}
                                onValueChange={(v) => v && setInviteRole(v as typeof inviteRole)}
                            >
                                <SelectTrigger>
                                    <SelectValue />
                                </SelectTrigger>
                                <SelectContent>
                                    {INVITE_ROLES.map((r) => (
                                        <SelectItem key={r.value} value={r.value}>
                                            {r.label}
                                        </SelectItem>
                                    ))}
                                </SelectContent>
                            </Select>
                        </div>
                    </div>
                    <DialogFooter>
                        <Button variant="outline" onClick={() => setInviteOpen(false)}>
                            Cancel
                        </Button>
                        <Button onClick={() => void sendInvite()} disabled={!inviteEmail.trim() || invitePending}>
                            <RiMailSendLine className="mr-1.5 size-4" />
                            {invitePending ? "Sending…" : "Send Invite"}
                        </Button>
                    </DialogFooter>
                </DialogContent>
            </Dialog>

            <Dialog open={!!removeTarget} onOpenChange={() => setRemoveTarget(null)}>
                <DialogContent className="sm:max-w-md">
                    <DialogHeader>
                        <DialogTitle className="text-destructive">Remove member</DialogTitle>
                        <DialogDescription>
                            Remove <span className="font-semibold text-foreground">{removeTarget?.email}</span> from
                            this organization? They&apos;ll lose access immediately.
                        </DialogDescription>
                    </DialogHeader>
                    <DialogFooter>
                        <Button variant="outline" onClick={() => setRemoveTarget(null)}>
                            Cancel
                        </Button>
                        <Button variant="destructive" onClick={() => void confirmRemove()}>
                            Remove
                        </Button>
                    </DialogFooter>
                </DialogContent>
            </Dialog>
        </div>
    )
}
