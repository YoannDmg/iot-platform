import { useDevices } from '@/hooks/useDevices';
import { StatsCard } from '@/components/stats-card';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Skeleton } from '@/components/ui/skeleton';
import { Activity, Cpu, AlertCircle, Zap } from 'lucide-react';
import { DeviceStatus } from '@/components/DeviceStatus';
import { Link } from 'react-router-dom';
import { formatTimestamp } from '@/lib/utils';

export function Dashboard() {
  const { data: devices, isLoading } = useDevices();

  const stats = {
    total: devices?.length || 0,
    online: devices?.filter((d) => d.status === 'ONLINE').length || 0,
    offline: devices?.filter((d) => d.status === 'OFFLINE').length || 0,
    alerts: devices?.filter((d) => d.status === 'ERROR').length || 0,
  };

  const uptime = stats.total > 0 ? ((stats.online / stats.total) * 100).toFixed(1) : '0';

  if (isLoading) {
    return (
      <div className="flex-1 space-y-4 p-8 pt-6">
        <div className="flex items-center justify-between">
          <Skeleton className="h-8 w-48" />
        </div>
        <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
          {[1, 2, 3, 4].map((i) => (
            <Skeleton key={i} className="h-32" />
          ))}
        </div>
      </div>
    );
  }

  return (
    <div className="flex-1 space-y-4 p-8 pt-6">
      <div className="flex items-center justify-between">
        <h2 className="text-3xl font-bold tracking-tight">Dashboard</h2>
      </div>

      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
        <StatsCard
          title="Total Devices"
          value={stats.total}
          description="All registered devices"
          icon={Cpu}
        />
        <StatsCard
          title="Online Devices"
          value={stats.online}
          description="Currently active"
          icon={Activity}
          trend={{ value: 12, isPositive: true }}
        />
        <StatsCard
          title="Uptime"
          value={`${uptime}%`}
          description="System availability"
          icon={Zap}
        />
        <StatsCard
          title="Alerts"
          value={stats.alerts}
          description="Devices with errors"
          icon={AlertCircle}
          trend={{ value: 8, isPositive: false }}
        />
      </div>

      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-7">
        <Card className="col-span-4">
          <CardHeader>
            <CardTitle>Recent Devices</CardTitle>
            <CardDescription>Latest device activity</CardDescription>
          </CardHeader>
          <CardContent>
            <div className="space-y-4">
              {devices?.slice(0, 5).map((device) => (
                <Link
                  key={device.id}
                  to={`/device/${device.id}`}
                  className="flex items-center justify-between rounded-lg border p-3 hover:bg-accent transition-colors"
                >
                  <div className="flex items-center gap-3">
                    <DeviceStatus status={device.status} />
                    <div>
                      <p className="font-medium">{device.name}</p>
                      <p className="text-sm text-muted-foreground">{device.type}</p>
                    </div>
                  </div>
                  <div className="text-right">
                    <p className="text-sm text-muted-foreground">
                      {formatTimestamp(device.lastSeen)}
                    </p>
                  </div>
                </Link>
              ))}
              {(!devices || devices.length === 0) && (
                <p className="text-center text-muted-foreground py-8">No devices found</p>
              )}
            </div>
          </CardContent>
        </Card>

        <Card className="col-span-3">
          <CardHeader>
            <CardTitle>Quick Stats</CardTitle>
            <CardDescription>Device distribution</CardDescription>
          </CardHeader>
          <CardContent>
            <div className="space-y-4">
              <div className="flex items-center justify-between">
                <div className="flex items-center gap-2">
                  <div className="h-3 w-3 rounded-full bg-green-500" />
                  <span className="text-sm">Online</span>
                </div>
                <span className="font-medium">{stats.online}</span>
              </div>
              <div className="flex items-center justify-between">
                <div className="flex items-center gap-2">
                  <div className="h-3 w-3 rounded-full bg-gray-400" />
                  <span className="text-sm">Offline</span>
                </div>
                <span className="font-medium">{stats.offline}</span>
              </div>
              <div className="flex items-center justify-between">
                <div className="flex items-center gap-2">
                  <div className="h-3 w-3 rounded-full bg-red-500" />
                  <span className="text-sm">Error</span>
                </div>
                <span className="font-medium">{stats.alerts}</span>
              </div>
            </div>
          </CardContent>
        </Card>
      </div>
    </div>
  );
}
