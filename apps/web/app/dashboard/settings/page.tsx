"use client"

import { useState, useTransition } from "react"
import { useRouter } from "next/navigation"
import { useQuery, useQueryClient } from "@tanstack/react-query"
import { toast } from "sonner"
import {
    RiCheckLine,
    RiDeleteBinLine,
    RiErrorWarningLine,
    RiImageAddLine,
    RiSave3Line,
    RiUploadCloud2Line,
} from "@remixicon/react"

import { Button } from "@/components/ui/button"
import {
    Card,
    CardContent,
    CardDescription,
    CardFooter,
    CardHeader,
    CardTitle,
} from "@/components/ui/card"
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
import {
    deleteOrgLogo,
    deleteOrganization,
    getOrgSettings,
    updateOrgSettings,
    uploadOrgLogo,
} from "@/lib/client-api/organizations"
import { queryKeys } from "@/react-query/query-keys"

export default function SettingsPage() {
    const router = useRouter()
    const queryClient = useQueryClient()
    const [isPending, startTransition] = useTransition()

    const settingsQuery = useQuery({
        queryKey: queryKeys.orgSettings(),
        queryFn: getOrgSettings,
    })
    const settings = settingsQuery.data

    // Organization Metadata State
    const [orgName, setOrgName] = useState(settings?.name ?? "")
    const [initialOrgName] = useState(settings?.name ?? "")

    // Logo Uploader State
    const [hasLogo, setHasLogo] = useState(settings?.has_logo ?? false)
    const [previewLogoUrl, setPreviewLogoUrl] = useState<string | null>(
        settings?.has_logo ? "/dashboard/settings/logo" : null
    )
    const [selectedFile, setSelectedFile] = useState<File | null>(null)

    // Delete Org Danger Zone Modal State
    const [deleteModalOpen, setDeleteModalOpen] = useState(false)
    const [confirmNameInput, setConfirmNameInput] = useState("")
    const [confirmPhraseInput, setConfirmPhraseInput] = useState("")

    // Handle Logo File Selection
    function handleFileSelect(e: React.ChangeEvent<HTMLInputElement>) {
        const file = e.target.files?.[0]
        if (!file) return

        if (file.size > 2 * 1024 * 1024) {
            toast.error("Logo image file size must be less than 2MB.")
            return
        }

        setSelectedFile(file)
        const localUrl = URL.createObjectURL(file)
        setPreviewLogoUrl(localUrl)
        setHasLogo(true)
    }

    // Handle Removing Logo
    function handleRemoveLogo() {
        startTransition(async () => {
            try {
                if (settings?.has_logo && !selectedFile) {
                    await deleteOrgLogo()
                }
            } catch (err: any) {
                toast.error(err.message || "Failed to delete logo.")
                return
            }
            setSelectedFile(null)
            setPreviewLogoUrl(null)
            setHasLogo(false)
            queryClient.invalidateQueries({ queryKey: queryKeys.orgSettings() })
            toast.success("Logo removed.")
        })
    }

    // Handle Saving Organization Metadata & Logo
    function handleSaveMetadata(e: React.FormEvent) {
        e.preventDefault()
        if (!orgName.trim()) {
            toast.error("Organization name cannot be empty.")
            return
        }

        startTransition(async () => {
            try {
                // 1. Update Name metadata
                if (orgName.trim() !== initialOrgName) {
                    await updateOrgSettings({ name: orgName.trim() })
                }

                // 2. Upload Logo File if selected
                if (selectedFile) {
                    const formData = new FormData()
                    formData.append("logo", selectedFile)
                    await uploadOrgLogo(formData)
                }

                queryClient.invalidateQueries({ queryKey: queryKeys.orgSettings() })
                toast.success("Organization metadata updated successfully!")
            } catch (err: any) {
                toast.error(err.message || "Failed to save changes.")
            }
        })
    }

    // Confirmation condition for Danger Zone Org Deletion
    const isDeleteConfirmed =
        confirmNameInput.trim() === initialOrgName.trim() &&
        confirmPhraseInput.trim().toLowerCase() === "delete organization"

    // Handle Delete Organization
    function handleConfirmDeleteOrg() {
        if (!isDeleteConfirmed) return

        startTransition(async () => {
            try {
                await deleteOrganization()
            } catch (err: any) {
                toast.error(err.message || "Failed to delete organization.")
                return
            }
            toast.success("Organization deleted successfully.")
            setDeleteModalOpen(false)
            router.push("/login")
        })
    }

    return (
        <div className="flex justify-center flex-1 w-full">
            <div className="flex flex-col gap-8 max-w-3xl">
                {/* Page Header */}
                <div>
                    <h1 className="font-heading text-2xl font-bold tracking-tight text-foreground sm:text-3xl">
                        Organization Settings
                    </h1>
                    <p className="mt-1 text-sm text-muted-foreground">
                        Manage your organization&apos;s identity, brand assets, and administrative settings.
                    </p>
                </div>

                {/* CARD 1: Organization Metadata Card */}
                <Card className="shadow-2xs">
                    <form onSubmit={handleSaveMetadata}>
                        <CardHeader>
                            <CardTitle className="text-lg font-semibold flex items-center gap-2">
                                <RiImageAddLine className="size-5 text-primary" />
                                <span>Organization Identity</span>
                            </CardTitle>
                            <CardDescription>
                                Update your organization name, slug, and emailed ID card brand logo.
                            </CardDescription>
                        </CardHeader>

                        <CardContent className="flex flex-col gap-6">
                            {/* Organization Name Field */}
                            <div className="flex flex-col gap-2 max-w-md">
                                <Label htmlFor="org-name" className="font-semibold text-xs text-foreground">
                                    Organization Name
                                </Label>
                                <Input
                                    id="org-name"
                                    value={orgName}
                                    onChange={(e) => setOrgName(e.target.value)}
                                    placeholder="Acme Corporation"
                                    className="font-medium"
                                />
                                <p className="text-xs text-muted-foreground">
                                    Slug: <code className="rounded bg-muted px-1.5 py-0.5 font-mono">{settings?.slug}</code>
                                </p>
                            </div>

                            {/* Logo File Uploader & Uploaded Image View */}
                            <div className="flex flex-col gap-2 border-t pt-5">
                                <Label className="font-semibold text-xs text-foreground">
                                    Organization Logo
                                </Label>
                                <p className="text-xs text-muted-foreground mb-1">
                                    Display on generated ID cards and emails. Recommended format: PNG or JPEG (Max 2MB).
                                </p>

                                {hasLogo && previewLogoUrl ? (
                                    /* Uploaded Image View Box */
                                    <div className="flex flex-col sm:flex-row items-start sm:items-center gap-4 rounded-lg border p-4 bg-muted/20">
                                        <div className="flex size-20 shrink-0 items-center justify-center rounded-lg border bg-background p-2 shadow-2xs">
                                            {/* eslint-disable-next-line @next/next/no-img-element */}
                                            <img
                                                src={previewLogoUrl}
                                                alt="Uploaded organization logo"
                                                className="max-h-full max-w-full object-contain"
                                            />
                                        </div>
                                        <div className="flex flex-col gap-1 flex-1">
                                            <span className="text-sm font-semibold text-foreground">
                                                Current Logo Uploaded
                                            </span>
                                            <span className="text-xs text-muted-foreground">
                                                {selectedFile ? selectedFile.name : "Active logo file image"}
                                            </span>
                                        </div>
                                        <Button
                                            type="button"
                                            variant="outline"
                                            size="sm"
                                            onClick={handleRemoveLogo}
                                            disabled={isPending}
                                            className="border-destructive/60 text-destructive hover:bg-destructive/10 gap-1.5 text-xs font-medium"
                                        >
                                            <RiDeleteBinLine className="size-3.5" />
                                            Remove Logo
                                        </Button>
                                    </div>
                                ) : (
                                    /* Drag & Drop / File Picker View Box */
                                    <label className="flex flex-col items-center justify-center w-full h-36 border-2 border-dashed rounded-lg cursor-pointer bg-muted/10 hover:bg-muted/30 transition-colors p-4 text-center group">
                                        <RiUploadCloud2Line className="size-8 text-muted-foreground group-hover:text-primary transition-colors mb-2" />
                                        <span className="text-sm font-semibold text-foreground">
                                            Click to select or drag logo image here
                                        </span>
                                        <span className="text-xs text-muted-foreground mt-1">
                                            Supports PNG, JPG, or WEBP (Up to 2MB)
                                        </span>
                                        <input
                                            type="file"
                                            accept="image/png,image/jpeg,image/webp"
                                            onChange={handleFileSelect}
                                            className="hidden"
                                        />
                                    </label>
                                )}
                            </div>
                        </CardContent>

                        <CardFooter className="border-t bg-muted/20 px-6 py-3.5 flex justify-end">
                            <Button type="submit" disabled={isPending} className="gap-2 font-semibold">
                                <RiSave3Line className="size-4" />
                                {isPending ? "Saving Changes…" : "Save Changes"}
                            </Button>
                        </CardFooter>
                    </form>
                </Card>

                {/* CARD 2: Danger Zone Card */}
                <Card className="border-destructive/40 bg-destructive/5 dark:bg-destructive/10 shadow-2xs">
                    <CardHeader>
                        <CardTitle className="text-lg font-semibold flex items-center gap-2 text-destructive">
                            <RiErrorWarningLine className="size-5 text-destructive" />
                            <span>Danger Zone</span>
                        </CardTitle>
                        <CardDescription className="text-destructive/90">
                            Permanently delete this organization, all associated events, members, and scanner devices. This action is irreversible.
                        </CardDescription>
                    </CardHeader>

                    <CardContent>
                        <p className="text-xs text-muted-foreground leading-relaxed">
                            Deleting your organization will immediately revoke access for all members, purge attendee records, and erase scanner device configurations permanently.
                        </p>
                    </CardContent>

                    <CardFooter className="border-t border-destructive/20 px-6 py-3.5 flex justify-end">
                        <Button
                            variant="destructive"
                            onClick={() => setDeleteModalOpen(true)}
                            className="gap-2 font-semibold"
                        >
                            <RiDeleteBinLine className="size-4" />
                            Delete Organization
                        </Button>
                    </CardFooter>
                </Card>

                {/* DANGER ZONE CONFIRMATION MODAL */}
                <Dialog open={deleteModalOpen} onOpenChange={setDeleteModalOpen}>
                    <DialogContent className="sm:max-w-md">
                        <DialogHeader>
                            <DialogTitle className="flex items-center gap-2 text-destructive">
                                <RiErrorWarningLine className="size-5 text-destructive" />
                                <span>Confirm Organization Deletion</span>
                            </DialogTitle>
                            <DialogDescription className="pt-2">
                                This action <strong className="text-foreground">cannot be undone</strong>. This will permanently delete{" "}
                                <strong className="text-foreground">{initialOrgName}</strong> and all associated data.
                            </DialogDescription>
                        </DialogHeader>

                        <div className="flex flex-col gap-4 py-3">
                            {/* Field 1: Type Organization Name */}
                            <div className="flex flex-col gap-1.5">
                                <Label className="text-xs text-muted-foreground">
                                    1. Type the organization name{" "}
                                    <code className="select-all font-bold text-foreground bg-muted px-1.5 py-0.5 rounded">
                                        {initialOrgName}
                                    </code>
                                </Label>
                                <Input
                                    value={confirmNameInput}
                                    onChange={(e) => setConfirmNameInput(e.target.value)}
                                    placeholder={initialOrgName}
                                />
                            </div>

                            {/* Field 2: Type "delete organization" */}
                            <div className="flex flex-col gap-1.5">
                                <Label className="text-xs text-muted-foreground">
                                    2. Type the confirmation phrase{" "}
                                    <code className="select-all font-bold text-foreground bg-muted px-1.5 py-0.5 rounded">
                                        delete organization
                                    </code>
                                </Label>
                                <Input
                                    value={confirmPhraseInput}
                                    onChange={(e) => setConfirmPhraseInput(e.target.value)}
                                    placeholder="delete organization"
                                />
                            </div>
                        </div>

                        <DialogFooter>
                            <Button
                                type="button"
                                variant="outline"
                                onClick={() => setDeleteModalOpen(false)}
                            >
                                Cancel
                            </Button>
                            <Button
                                variant="destructive"
                                disabled={!isDeleteConfirmed || isPending}
                                onClick={handleConfirmDeleteOrg}
                                className="gap-1.5"
                            >
                                {isPending ? (
                                    "Deleting…"
                                ) : (
                                    <>
                                        <RiCheckLine className="size-4" />
                                        Confirm Deletion
                                    </>
                                )}
                            </Button>
                        </DialogFooter>
                    </DialogContent>
                </Dialog>
            </div>
        </div>
    )
}