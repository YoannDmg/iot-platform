import { useParams, Link } from 'react-router-dom';
import { useDevice } from '@/hooks/useDevice';
import { DeviceStatus } from '@/components/DeviceStatus';
import { MetricCard } from '@/components/MetricCard';
import { parseMetadata, formatTimestamp } from '@/lib/utils';
import { Card } from '@/components/ui/card';
import { Skeleton } from '@/components/ui/skeleton';
import { Button } from '@/components/ui/button';
import { ArrowLeft } from 'lucide-react';

export function DeviceDetail() {
  const { id } = useParams<{ id: string }>();
  const { data: device, isLoading, error } = useDevice(id!);

  if (isLoading) {
    return (
      <div className="flex-1 space-y-4 p-8 pt-6">
        <Skeleton className="h-10 w-32" />
        <Skeleton className="h-48 w-full" />
        <Skeleton className="h-8 w-32" />
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
          {[1, 2, 3, 4].map((i) => (
            <Skeleton key={i} className="h-32" />
          ))}
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="flex-1 p-8 pt-6">
        <div className="text-lg text-destructive">Error loading device: {error.message}</div>
      </div>
    );
  }

  if (!device) {
    return (
      <div className="flex-1 p-8 pt-6">
        <div className="text-lg text-muted-foreground">Device not found</div>
      </div>
    );
  }

  const metrics = parseMetadata(device.metadata || []);

  return (
    <div className="flex-1 space-y-4 p-8 pt-6">
      <div>
        <Button variant="ghost" asChild className="mb-4">
          <Link to="/devices">
            <ArrowLeft className="mr-2 h-4 w-4" />
            Back to devices
          </Link>
        </Button>

        <Card className="p-6 mb-6">
          <div className="flex items-start justify-between">
            <div>
              <h1 className="text-3xl font-bold mb-2">{device.name}</h1>
              <p className="text-muted-foreground">{device.type}</p>
              <p className="text-sm text-muted-foreground mt-2">ID: {device.id}</p>
            </div>
            <DeviceStatus status={device.status} />
          </div>

          <div className="mt-4 text-sm text-muted-foreground">
            Last seen: {formatTimestamp(device.lastSeen)}
          </div>
        </Card>

        <h2 className="text-2xl font-bold mb-4">Metrics</h2>

        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
          {metrics.cpuPercent !== undefined && (
            <MetricCard
              label="CPU Usage"
              value={metrics.cpuPercent.toFixed(1)}
              unit="%"
              icon="💻"
            />
          )}

          {metrics.memoryUsedGB !== undefined && (
            <MetricCard
              label="Memory Used"
              value={metrics.memoryUsedGB.toFixed(2)}
              unit="GB"
              icon="🧠"
            />
          )}

          {metrics.diskUsedGB !== undefined && (
            <MetricCard
              label="Disk Used"
              value={metrics.diskUsedGB.toFixed(2)}
              unit="GB"
              icon="💾"
            />
          )}

          {metrics.batteryLevel !== undefined && (
            <MetricCard
              label="Battery"
              value={metrics.batteryLevel}
              unit="%"
              icon="🔋"
            />
          )}

          {metrics.networkUpMB !== undefined && (
            <MetricCard
              label="Network Upload"
              value={metrics.networkUpMB.toFixed(2)}
              unit="MB"
              icon="⬆️"
            />
          )}

          {metrics.networkDownMB !== undefined && (
            <MetricCard
              label="Network Download"
              value={metrics.networkDownMB.toFixed(2)}
              unit="MB"
              icon="⬇️"
            />
          )}

          {metrics.processCount !== undefined && (
            <MetricCard
              label="Active Processes"
              value={metrics.processCount}
              icon="⚙️"
            />
          )}
        </div>

        {(!device.metadata || device.metadata.length === 0) && (
          <div className="text-center py-12">
            <p className="text-muted-foreground">
              No metrics available yet. Waiting for device to send data...
            </p>
          </div>
        )}
      </div>
    </div>
  );
}
