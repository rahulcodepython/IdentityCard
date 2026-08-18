import { getOrgSettings } from "@/lib/api/organizations"

import { LogoUploadForm } from "./logo-upload-form"

export default async function SettingsPage() {
  const settings = await getOrgSettings()

  return (
    <div className="flex flex-col gap-6">
      <h1 className="font-heading text-xl font-medium">
        Organization settings
      </h1>
      <p className="text-sm text-muted-foreground">
        {settings.name} · {settings.slug}
      </p>

      <div className="flex flex-col gap-3">
        <h2 className="text-sm font-medium">Logo</h2>
        {settings.has_logo && (
          // eslint-disable-next-line @next/next/no-img-element -- proxied via our own dynamic route, not an optimizable static asset
          <img
            src="/dashboard/settings/logo"
            alt="Organization logo"
            className="h-16 w-auto rounded border object-contain p-2"
          />
        )}
        <p className="text-xs text-muted-foreground">
          Shown on emailed ID cards. PNG or JPEG, up to 2MB.
        </p>
        <LogoUploadForm />
      </div>
    </div>
  )
}
