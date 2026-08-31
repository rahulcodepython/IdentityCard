"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { toast } from "sonner";
import {
    RiCheckLine,
    RiDeleteBinLine,
    RiErrorWarningLine,
    RiImageAddLine,
    RiSave3Line,
    RiUploadCloud2Line,
} from "@remixicon/react";

import { Button } from "@/components/ui/button";
import {
    Card,
    CardContent,
    CardDescription,
    CardFooter,
    CardHeader,
    CardTitle,
} from "@/components/ui/card";
import {
    Dialog,
    DialogContent,
    DialogDescription,
    DialogFooter,
    DialogHeader,
    DialogTitle,
} from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
    useDeleteLogoMutation,
    useDeleteOrganizationMutation,
    useOrgSettingsQuery,
    useUpdateOrgSettingsMutation,
    useUploadLogoMutation,
} from "@/query-hooks/organizations.api";

export default function SettingsPage() {
    const router = useRouter();

    const { data: settings } = useOrgSettingsQuery();
    const updateSettingsMutation = useUpdateOrgSettingsMutation();
    const uploadLogoMutation = useUploadLogoMutation();
    const deleteLogoMutation = useDeleteLogoMutation();
    const deleteOrgMutation = useDeleteOrganizationMutation();

    const [orgName, setOrgName] = useState(settings?.name ?? "");
    const initialOrgName = settings?.name ?? "";

    const [hasLogo, setHasLogo] = useState(settings?.has_logo ?? false);
    const [previewLogoUrl, setPreviewLogoUrl] = useState<string | null>(
        settings?.has_logo ? "/dashboard/settings/logo" : null
    );
    const [selectedFile, setSelectedFile] = useState<File | null>(null);

    const [deleteModalOpen, setDeleteModalOpen] = useState(false);
    const [confirmNameInput, setConfirmNameInput] = useState("");
    const [confirmPhraseInput, setConfirmPhraseInput] = useState("");

    const isPending =
        updateSettingsMutation.isPending ||
        uploadLogoMutation.isPending ||
        deleteLogoMutation.isPending ||
        deleteOrgMutation.isPending;

    function handleFileSelect(e: React.ChangeEvent<HTMLInputElement>) {
        const file = e.target.files?.[0];
        if (!file) return;

        if (file.size > 2 * 1024 * 1024) {
            toast.error("Logo image file size must be less than 2MB.");
            return;
        }

        setSelectedFile(file);
        const localUrl = URL.createObjectURL(file);
        setPreviewLogoUrl(localUrl);
        setHasLogo(true);
    }

    async function handleRemoveLogo() {
        if (settings?.has_logo && !selectedFile) {
            const res = await deleteLogoMutation.execute();
            if (res === null && deleteLogoMutation.error) {
                toast.error(deleteLogoMutation.error.message || "Failed to delete logo.");
                return;
            }
        }
        setSelectedFile(null);
        setPreviewLogoUrl(null);
        setHasLogo(false);
        toast.success("Logo removed.");
    }

    async function handleSaveMetadata(e: React.FormEvent) {
        e.preventDefault();
        const cleanName = orgName || initialOrgName;
        if (!cleanName.trim()) {
            toast.error("Organization name cannot be empty.");
            return;
        }

        if (cleanName.trim() !== initialOrgName) {
            const res = await updateSettingsMutation.execute({
                organization_name: cleanName.trim(),
            });
            if (!res && updateSettingsMutation.error) {
                toast.error(updateSettingsMutation.error.message || "Failed to update name.");
                return;
            }
        }

        if (selectedFile) {
            const formData = new FormData();
            formData.append("logo", selectedFile);
            const res = await uploadLogoMutation.execute(formData);
            if (res === null && uploadLogoMutation.error) {
                toast.error(uploadLogoMutation.error.message || "Failed to upload logo.");
                return;
            }
        }

        toast.success("Organization metadata updated successfully!");
    }

    const isDeleteConfirmed =
        confirmNameInput.trim() === initialOrgName.trim() &&
        confirmPhraseInput.trim().toLowerCase() === "delete organization";

    async function handleConfirmDeleteOrg() {
        if (!isDeleteConfirmed) return;

        const res = await deleteOrgMutation.execute();
        if (res !== null) {
            toast.success("Organization deleted successfully.");
            setDeleteModalOpen(false);
            router.push("/login");
        } else if (deleteOrgMutation.error) {
            toast.error(deleteOrgMutation.error.message || "Failed to delete organization.");
        }
    }

    return (
        <div className="flex justify-center flex-1 w-full">
            <div className="flex flex-col gap-8 max-w-3xl">
                <div>
                    <h1 className="font-heading text-2xl font-bold tracking-tight text-foreground sm:text-3xl">
                        Organization Settings
                    </h1>
                    <p className="mt-1 text-sm text-muted-foreground">
                        Manage your organization&apos;s identity, brand assets, and administrative settings.
                    </p>
                </div>

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
                            <div className="flex flex-col gap-2 max-w-md">
                                <Label htmlFor="org-name" className="font-semibold text-xs text-foreground">
                                    Organization Name
                                </Label>
                                <Input
                                    id="org-name"
                                    defaultValue={initialOrgName}
                                    onChange={(e) => setOrgName(e.target.value)}
                                    placeholder="Acme Corporation"
                                    className="font-medium"
                                />
                                <p className="text-xs text-muted-foreground">
                                    Slug: <code className="rounded bg-muted px-1.5 py-0.5 font-mono">{settings?.slug}</code>
                                </p>
                            </div>

                            <div className="flex flex-col gap-2 border-t pt-5">
                                <Label className="font-semibold text-xs text-foreground">
                                    Organization Logo
                                </Label>
                                <p className="text-xs text-muted-foreground mb-1">
                                    Display on generated ID cards and emails. Recommended format: PNG or JPEG (Max 2MB).
                                </p>

                                {hasLogo && previewLogoUrl ? (
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
    );
}