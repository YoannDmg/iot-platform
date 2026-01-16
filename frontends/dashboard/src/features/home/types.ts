// Types spécifiques à la page Overview
// Les types Device et DeviceStatus sont dans shared/types

export { type Device, type DeviceStatus } from "@/shared/types"

export interface Stats {
  totalDevices: number
  onlineDevices: number
  offlineDevices: number
  errorDevices: number
}

export interface ActivityItem {
  id: string
  deviceId: string
  deviceName: string
  deviceType: string
  status: import("@/shared/types").DeviceStatus
  action: string
  timestamp: string
}

export interface Alert {
  id: string
  deviceId: string
  deviceName: string
  type: "warning" | "error" | "info"
  message: string
  timestamp: string
}
