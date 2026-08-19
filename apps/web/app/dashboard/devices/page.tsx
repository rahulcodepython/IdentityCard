import { listDevices } from "@/lib/api/devices"

import { DevicesClient } from "./devices-client"

export default async function DevicesPage() {
  const devices = await listDevices()

  return <DevicesClient initialDevices={devices} />
}
