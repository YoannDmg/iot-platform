import { Link } from 'react-router-dom';
import { useDevices } from '@/hooks/useDevices';
import { DeviceStatus } from '@/components/DeviceStatus';
import { formatTimestamp } from '@/lib/utils';
import { Card } from '@/components/ui/card';
import { Skeleton } from '@/components/ui/skeleton';

export function DeviceList() {
  const { data: devices, isLoading, error } = useDevices();

  if (isLoading) {
    return (
      <div className="flex-1 space-y-4 p-8 pt-6">
        <Skeleton className="h-8 w-48" />
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
          {[1, 2, 3, 4, 5, 6].map((i) => (
            <Skeleton key={i} className="h-32" />
          ))}
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="flex-1 p-8 pt-6">
        <div className="text-lg text-destructive">Error loading devices: {error.message}</div>
      </div>
    );
  }

  return (
    <div className="flex-1 space-y-4 p-8 pt-6">
      <h1 className="text-3xl font-bold tracking-tight">Devices</h1>

      {devices && devices.length === 0 ? (
        <div className="text-center py-12">
          <p className="text-muted-foreground">No devices found</p>
        </div>
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
          {devices?.map((device) => (
            <Link key={device.id} to={`/device/${device.id}`}>
              <Card className="p-6 hover:shadow-lg transition-shadow cursor-pointer">
                <div className="flex items-start justify-between mb-4">
                  <div>
                    <h2 className="text-xl font-semibold">{device.name}</h2>
                    <p className="text-sm text-muted-foreground">{device.type}</p>
                  </div>
                  <DeviceStatus status={device.status} />
                </div>
                <div className="text-sm text-muted-foreground">
                  Last seen: {formatTimestamp(device.lastSeen)}
                </div>
              </Card>
            </Link>
          ))}
        </div>
      )}
    </div>
  );
}
