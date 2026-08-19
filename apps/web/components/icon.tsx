import type { ComponentProps } from "react"
import {
  type RemixiconComponentType,
  RiArrowDownLine,
  RiArrowDownSLine,
  RiArrowRightSLine,
  RiArrowUpSLine,
  RiBankCardLine,
  RiBuilding2Line,
  RiCalendarEventLine,
  RiCloseLine,
  RiComputerLine,
  RiDashboardLine,
  RiErrorWarningLine,
  RiExpandUpDownLine,
  RiLoader4Line,
  RiLogoutBoxRLine,
  RiMoonLine,
  RiSettings3Line,
  RiSmartphoneLine,
  RiSunLine,
  RiUser3Line,
} from "@remixicon/react"

// Registry wrapping @remixicon/react (already the icon lib used throughout
// the app) so call sites import <Icon name="..." /> instead of a vendor
// icon component directly — swapping icon libraries later touches only
// this file. Seeded with icons already in use as of Phase 3; Phase 4 call
// sites extend this map as they convert.
const iconRegistry = {
  "arrow-down": RiArrowDownLine,
  "arrow-down-s": RiArrowDownSLine,
  "arrow-right-s": RiArrowRightSLine,
  "arrow-up-s": RiArrowUpSLine,
  "bank-card": RiBankCardLine,
  building: RiBuilding2Line,
  "calendar-event": RiCalendarEventLine,
  close: RiCloseLine,
  computer: RiComputerLine,
  dashboard: RiDashboardLine,
  "error-warning": RiErrorWarningLine,
  "expand-up-down": RiExpandUpDownLine,
  loader: RiLoader4Line,
  logout: RiLogoutBoxRLine,
  moon: RiMoonLine,
  settings: RiSettings3Line,
  smartphone: RiSmartphoneLine,
  sun: RiSunLine,
  user: RiUser3Line,
} as const satisfies Record<string, RemixiconComponentType>

export type IconName = keyof typeof iconRegistry

export function Icon({
  name,
  ...props
}: { name: IconName } & ComponentProps<RemixiconComponentType>) {
  const Component = iconRegistry[name]
  return <Component {...props} />
}
