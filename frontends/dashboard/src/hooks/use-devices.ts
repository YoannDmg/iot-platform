import { useQuery, useMutation, useSubscription } from "@apollo/client/react"
import {
  GET_DEVICES,
  GET_DEVICE,
  GET_STATS,
  CREATE_DEVICE,
  UPDATE_DEVICE,
  DELETE_DEVICE,
  DEVICE_UPDATED,
  type GetDevicesResponse,
  type GetDevicesVariables,
  type GetDeviceResponse,
  type GetDeviceVariables,
  type GetStatsResponse,
  type CreateDeviceResponse,
  type CreateDeviceVariables,
  type UpdateDeviceResponse,
  type UpdateDeviceVariables,
  type DeleteDeviceResponse,
  type DeleteDeviceVariables,
} from "@/graphql"
import type { Device } from "@/types/device"

export function useDevices(variables?: GetDevicesVariables) {
  const { data, loading, error, refetch } = useQuery<
    GetDevicesResponse,
    GetDevicesVariables
  >(GET_DEVICES, {
    variables: {
      page: 1,
      pageSize: 20,
      ...variables,
    },
  })

  return {
    devices: data?.devices.devices ?? [],
    total: data?.devices.total ?? 0,
    page: data?.devices.page ?? 1,
    pageSize: data?.devices.pageSize ?? 20,
    loading,
    error,
    refetch,
  }
}

export function useDevice(id: string) {
  const { data, loading, error, refetch } = useQuery<
    GetDeviceResponse,
    GetDeviceVariables
  >(GET_DEVICE, {
    variables: { id },
    skip: !id,
  })

  return {
    device: data?.device ?? null,
    loading,
    error,
    refetch,
  }
}

export function useStats() {
  const { data, loading, error, refetch } = useQuery<GetStatsResponse>(GET_STATS)

  return {
    stats: data?.stats ?? null,
    loading,
    error,
    refetch,
  }
}

export function useCreateDevice() {
  const [createDevice, { loading, error }] = useMutation<
    CreateDeviceResponse,
    CreateDeviceVariables
  >(CREATE_DEVICE, {
    refetchQueries: [GET_DEVICES, GET_STATS],
  })

  const create = async (input: CreateDeviceVariables["input"]) => {
    const result = await createDevice({ variables: { input } })
    return result.data?.createDevice
  }

  return {
    createDevice: create,
    loading,
    error,
  }
}

export function useUpdateDevice() {
  const [updateDevice, { loading, error }] = useMutation<
    UpdateDeviceResponse,
    UpdateDeviceVariables
  >(UPDATE_DEVICE)

  const update = async (input: UpdateDeviceVariables["input"]) => {
    const result = await updateDevice({ variables: { input } })
    return result.data?.updateDevice
  }

  return {
    updateDevice: update,
    loading,
    error,
  }
}

export function useDeleteDevice() {
  const [deleteDevice, { loading, error }] = useMutation<
    DeleteDeviceResponse,
    DeleteDeviceVariables
  >(DELETE_DEVICE, {
    refetchQueries: [GET_DEVICES, GET_STATS],
  })

  const remove = async (id: string) => {
    const result = await deleteDevice({ variables: { id } })
    return result.data?.deleteDevice
  }

  return {
    deleteDevice: remove,
    loading,
    error,
  }
}

export function useDeviceUpdates(onUpdate?: (device: Device) => void) {
  const { data, loading, error } = useSubscription<{ deviceUpdated: Device }>(
    DEVICE_UPDATED,
    {
      onData: ({ data: subscriptionData }) => {
        if (subscriptionData.data?.deviceUpdated && onUpdate) {
          onUpdate(subscriptionData.data.deviceUpdated)
        }
      },
    }
  )

  return {
    device: data?.deviceUpdated ?? null,
    loading,
    error,
  }
}
