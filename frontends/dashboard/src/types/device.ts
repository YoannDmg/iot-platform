export interface Device {
  id: string;
  name: string;
  type: string;
  status: 'UNKNOWN' | 'ONLINE' | 'OFFLINE' | 'ERROR' | 'MAINTENANCE';
  metadata?: DeviceMetadata[];
  createdAt: number;
  lastSeen: number;
}

export interface DeviceMetadata {
  key: string;
  value: string;
}

export interface DeviceConnection {
  devices: Device[];
  total: number;
  page: number;
  pageSize: number;
}

export interface DeviceMetrics {
  cpuPercent?: number;
  memoryUsedGB?: number;
  diskUsedGB?: number;
  networkUpMB?: number;
  networkDownMB?: number;
  batteryLevel?: number;
  processCount?: number;
}
