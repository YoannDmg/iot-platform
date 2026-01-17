// Types partagés pour les devices 
export type DeviceStatus = "UNKNOWN" | "ONLINE" | "OFFLINE" | "ERROR" | "MAINTENANCE"

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

export interface DeviceConnection {
  devices: Device[]
  total: number
  page: number
  pageSize: number
}

export interface CreateDeviceInput {
  name: string
  type: string
  metadata?: DeviceMetadata[]
}

export interface UpdateDeviceInput {
  id: string
  name?: string
  status?: DeviceStatus
  metadata?: DeviceMetadata[]
}

export interface DeleteResult {
  success: boolean
  message: string
}

export interface Stats {
  totalDevices: number
  onlineDevices: number
  offlineDevices: number
  errorDevices: number
}

