import { getOrgSettings } from "@/lib/api/organizations"

import { SettingsClient } from "./settings-client"

export default async function SettingsPage() {
  const settings = await getOrgSettings()

  return <SettingsClient settings={settings} />
}
