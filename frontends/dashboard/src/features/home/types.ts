// Types pour la page Overview

export type DeviceStatus = "ONLINE" | "OFFLINE" | "ERROR" | "MAINTENANCE" | "UNKNOWN"

export interface Device {
  id: string
  name: string
  type: string
  status: DeviceStatus
  createdAt: number
  lastSeen: number
  metadata: Record<string, string>
}

export interface Stats {
  totalDevices: number
  onlineDevices: number
  offlineDevices: number
  errorDevices: number
}

export interface ActivityItem {
  id: string
  device: Device
  action: string
  timestamp: number
}

export interface Alert {
  id: string
  deviceId: string
  deviceName: string
  type: "warning" | "error" | "info"
  message: string
  timestamp: number
}
