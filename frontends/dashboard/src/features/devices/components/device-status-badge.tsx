import { Badge } from "@/components/ui/badge"
import type { DeviceStatus } from "@/types/device"
import { cn } from "@/lib/utils"

const deviceStatusConfig: Record<
  DeviceStatus,
  { color: string; bgColor: string; label: string }
> = {
  ONLINE: {
    color: "text-green-600 dark:text-green-400",
    bgColor: "bg-green-500",
    label: "Online",
  },
  OFFLINE: {
    color: "text-red-600 dark:text-red-400",
    bgColor: "bg-red-500",
    label: "Offline",
  },
  ERROR: {
    color: "text-red-600 dark:text-red-400",
    bgColor: "bg-red-500",
    label: "Error",
  },
  MAINTENANCE: {
    color: "text-yellow-600 dark:text-yellow-400",
    bgColor: "bg-yellow-500",
    label: "Maintenance",
  },
  UNKNOWN: {
    color: "text-gray-600 dark:text-gray-400",
    bgColor: "bg-gray-500",
    label: "Unknown",
  },
}

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
