import { useParams, Link } from 'react-router-dom';
import { useDevice } from '../hooks/useDevice';
import { DeviceStatus } from '../components/DeviceStatus';
import { MetricCard } from '../components/MetricCard';
import { parseMetadata, formatTimestamp } from '../lib/utils';

export function DeviceDetail() {
  const { id } = useParams<{ id: string }>();
  const { data: device, isLoading, error } = useDevice(id!);

  if (isLoading) {
    return (
      <div className="flex items-center justify-center h-screen">
        <div className="text-lg">Loading device...</div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="flex items-center justify-center h-screen">
        <div className="text-lg text-red-500">Error loading device: {error.message}</div>
      </div>
    );
  }

  if (!device) {
    return (
      <div className="flex items-center justify-center h-screen">
        <div className="text-lg text-gray-500">Device not found</div>
      </div>
    );
  }

  const metrics = parseMetadata(device.metadata || []);

  return (
    <div className="min-h-screen bg-gray-50 dark:bg-gray-900">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        <Link
          to="/"
          className="text-blue-600 dark:text-blue-400 hover:underline mb-4 inline-block"
        >
          ← Back to devices
        </Link>

        <div className="bg-white dark:bg-gray-800 rounded-lg shadow-lg p-6 mb-6">
          <div className="flex items-start justify-between">
            <div>
              <h1 className="text-3xl font-bold text-gray-900 dark:text-white mb-2">
                {device.name}
              </h1>
              <p className="text-gray-600 dark:text-gray-400">{device.type}</p>
              <p className="text-sm text-gray-500 dark:text-gray-500 mt-2">
                ID: {device.id}
              </p>
            </div>
            <DeviceStatus status={device.status} />
          </div>

          <div className="mt-4 text-sm text-gray-500 dark:text-gray-400">
            Last seen: {formatTimestamp(device.lastSeen)}
          </div>
        </div>

        <h2 className="text-2xl font-bold text-gray-900 dark:text-white mb-4">
          Metrics
        </h2>

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
            <p className="text-gray-500 dark:text-gray-400">
              No metrics available yet. Waiting for device to send data...
            </p>
          </div>
        )}
      </div>
    </div>
  );
}
