import { gql } from "@apollo/client"

// ============================================
// Fragments
// ============================================

export const DEVICE_FRAGMENT = gql`
  fragment DeviceFields on Device {
    id
    name
    type
    status
    createdAt
    lastSeen
    metadata {
      key
      value
    }
  }
`

export const USER_FRAGMENT = gql`
  fragment UserFields on User {
    id
    email
    name
    role
    createdAt
    lastLogin
    isActive
  }
`

// ============================================
// Queries
// ============================================

export const GET_ME = gql`
  query GetMe {
    me {
      ...UserFields
    }
  }
  ${USER_FRAGMENT}
`

export const GET_DEVICE = gql`
  query GetDevice($id: ID!) {
    device(id: $id) {
      ...DeviceFields
    }
  }
  ${DEVICE_FRAGMENT}
`

export const GET_DEVICES = gql`
  query GetDevices($page: Int, $pageSize: Int, $type: String, $status: DeviceStatus) {
    devices(page: $page, pageSize: $pageSize, type: $type, status: $status) {
      devices {
        ...DeviceFields
      }
      total
      page
      pageSize
    }
  }
  ${DEVICE_FRAGMENT}
`

export const GET_STATS = gql`
  query GetStats {
    stats {
      totalDevices
      onlineDevices
      offlineDevices
      errorDevices
    }
  }
`

// ============================================
// Mutations
// ============================================

export const LOGIN = gql`
  mutation Login($input: LoginInput!) {
    login(input: $input) {
      token
      user {
        ...UserFields
      }
    }
  }
  ${USER_FRAGMENT}
`

export const REGISTER = gql`
  mutation Register($input: RegisterInput!) {
    register(input: $input) {
      token
      user {
        ...UserFields
      }
    }
  }
  ${USER_FRAGMENT}
`

export const CREATE_DEVICE = gql`
  mutation CreateDevice($input: CreateDeviceInput!) {
    createDevice(input: $input) {
      ...DeviceFields
    }
  }
  ${DEVICE_FRAGMENT}
`

export const UPDATE_DEVICE = gql`
  mutation UpdateDevice($input: UpdateDeviceInput!) {
    updateDevice(input: $input) {
      ...DeviceFields
    }
  }
  ${DEVICE_FRAGMENT}
`

export const DELETE_DEVICE = gql`
  mutation DeleteDevice($id: ID!) {
    deleteDevice(id: $id) {
      success
      message
    }
  }
`

// ============================================
// Subscriptions
// ============================================

export const DEVICE_UPDATED = gql`
  subscription DeviceUpdated {
    deviceUpdated {
      ...DeviceFields
    }
  }
  ${DEVICE_FRAGMENT}
`
