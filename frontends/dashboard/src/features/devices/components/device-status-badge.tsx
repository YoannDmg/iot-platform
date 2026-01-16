import { Badge } from "@/components/ui/badge"
import { type DeviceStatus, deviceStatusConfig } from "@/shared/types"
import { cn } from "@/lib/utils"

interface DeviceStatusBadgeProps {
  status: DeviceStatus
  className?: string
}

export function DeviceStatusBadge({ status, className }: DeviceStatusBadgeProps) {
  const config = deviceStatusConfig[status]

  return (
    <Badge variant="outline" className={cn("gap-1.5", className)}>
      <span className={cn("h-2 w-2 rounded-full", config.bgColor)} />
      {config.label}
    </Badge>
  )
}
