import { clsx, type ClassValue } from "clsx"
import { twMerge } from "tailwind-merge"
import type { DeviceMetadata, DeviceMetrics } from '../types/device';

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs))
}

export function parseMetadata(metadata: DeviceMetadata[]): DeviceMetrics {
  const metrics: DeviceMetrics = {};

  metadata.forEach((item) => {
    const value = parseFloat(item.value);
    if (!isNaN(value)) {
      switch (item.key) {
        case 'cpu_percent':
          metrics.cpuPercent = value;
          break;
        case 'memory_used_gb':
        case 'memory_gb':
          metrics.memoryUsedGB = value;
          break;
        case 'disk_used_gb':
        case 'disk_gb':
          metrics.diskUsedGB = value;
          break;
        case 'network_up_mb':
          metrics.networkUpMB = value;
          break;
        case 'network_down_mb':
          metrics.networkDownMB = value;
          break;
        case 'battery_level':
          metrics.batteryLevel = value;
          break;
        case 'process_count':
          metrics.processCount = value;
          break;
      }
    }
  });

  return metrics;
}

export function formatTimestamp(timestamp: number): string {
  // Convert Unix timestamp (seconds) to milliseconds
  const date = new Date(timestamp * 1000);
  const now = new Date();
  const diff = now.getTime() - date.getTime();
  const seconds = Math.floor(diff / 1000);

  if (seconds < 60) return `${seconds}s ago`;
  if (seconds < 3600) return `${Math.floor(seconds / 60)}m ago`;
  if (seconds < 86400) return `${Math.floor(seconds / 3600)}h ago`;
  return date.toLocaleDateString();
}
