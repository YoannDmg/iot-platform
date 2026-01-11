import { useQuery } from '@tanstack/react-query';
import { graphqlClient } from '../lib/graphql-client';
import { GET_DEVICES } from '../lib/queries';
import type { DeviceConnection } from '../types/device';

interface GetDevicesResponse {
  devices: DeviceConnection;
}

export function useDevices(page = 1, pageSize = 100) {
  return useQuery({
    queryKey: ['devices', page, pageSize],
    queryFn: async () => {
      const data = await graphqlClient.request<GetDevicesResponse>(GET_DEVICES, { page, pageSize });
      return data.devices.devices; // Return the devices array from the connection
    },
    refetchInterval: 10000, // Refresh every 10 seconds
  });
}
