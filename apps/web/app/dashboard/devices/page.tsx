import { listDevices } from "@/lib/api/devices"

import { CreateDeviceForm } from "./create-device-form"
import { DeviceRow } from "./device-row"

export default async function DevicesPage() {
  const devices = await listDevices()

  return (
    <div className="flex flex-col gap-6">
      <h1 className="font-heading text-xl font-medium">Scanner devices</h1>
      <p className="text-sm text-muted-foreground">
        Add a device, then on that device visit <code>/pair</code> and enter the
        code shown below before it expires.
      </p>

      <CreateDeviceForm />

      {devices.length === 0 ? (
        <p className="text-sm text-muted-foreground">No devices yet.</p>
      ) : (
        <div className="flex flex-col gap-2">
          {devices.map((device) => (
            <DeviceRow key={device.id} device={device} />
          ))}
        </div>
      )}
    </div>
  )
}
