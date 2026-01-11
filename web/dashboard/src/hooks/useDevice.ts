import { useQuery } from '@tanstack/react-query';
import { graphqlClient } from '../lib/graphql-client';
import { GET_DEVICE } from '../lib/queries';
import type { Device } from '../types/device';

interface GetDeviceResponse {
  device: Device;
}

export function useDevice(deviceId: string) {
  return useQuery({
    queryKey: ['device', deviceId],
    queryFn: async () => {
      const data = await graphqlClient.request<GetDeviceResponse>(GET_DEVICE, { id: deviceId });
      return data.device;
    },
    refetchInterval: 5000, // Refresh every 5 seconds
  });
}
