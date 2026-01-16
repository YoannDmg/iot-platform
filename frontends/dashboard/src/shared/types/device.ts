// Types partagés pour les devices

export type DeviceStatus = "ONLINE" | "OFFLINE" | "ERROR" | "MAINTENANCE" | "UNKNOWN"

export interface DeviceMetadata {
  key: string
  value: string
}

export interface Device {
  id: string
  name: string
  type: string
  status: DeviceStatus
  createdAt: number
  lastSeen: number
  metadata: DeviceMetadata[]
}

export const deviceStatusConfig: Record<
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
