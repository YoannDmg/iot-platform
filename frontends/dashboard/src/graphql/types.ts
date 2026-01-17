// Types pour les réponses et variables API GraphQL

import type {
  Device,
  DeviceConnection,
  DeviceStatus,
  CreateDeviceInput,
  UpdateDeviceInput,
  DeleteResult,
  Stats,
} from "@/types/device"
import type { User, AuthPayload, LoginInput, RegisterInput } from "@/types/user"

// Query response types
export interface GetMeResponse {
  me: User | null
}

export interface GetDeviceResponse {
  device: Device | null
}

export interface GetDevicesResponse {
  devices: DeviceConnection
}

export interface GetStatsResponse {
  stats: Stats
}

// Mutation response types
export interface LoginResponse {
  login: AuthPayload
}

export interface RegisterResponse {
  register: AuthPayload
}

export interface CreateDeviceResponse {
  createDevice: Device
}

export interface UpdateDeviceResponse {
  updateDevice: Device
}

export interface DeleteDeviceResponse {
  deleteDevice: DeleteResult
}

// Query variables types
export interface GetDeviceVariables {
  id: string
}

export interface GetDevicesVariables {
  page?: number
  pageSize?: number
  type?: string
  status?: DeviceStatus
}

export interface LoginVariables {
  input: LoginInput
}

export interface RegisterVariables {
  input: RegisterInput
}

export interface CreateDeviceVariables {
  input: CreateDeviceInput
}

export interface UpdateDeviceVariables {
  input: UpdateDeviceInput
}

export interface DeleteDeviceVariables {
  id: string
}
