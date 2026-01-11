import { Link } from 'react-router-dom';
import { useDevices } from '../hooks/useDevices';
import { DeviceStatus } from '../components/DeviceStatus';
import { formatTimestamp } from '../lib/utils';

export function DeviceList() {
  const { data: devices, isLoading, error } = useDevices();

  if (isLoading) {
    return (
      <div className="flex items-center justify-center h-screen">
        <div className="text-lg">Loading devices...</div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="flex items-center justify-center h-screen">
        <div className="text-lg text-red-500">Error loading devices: {error.message}</div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-gray-50 dark:bg-gray-900">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        <h1 className="text-3xl font-bold text-gray-900 dark:text-white mb-8">
          IoT Devices
        </h1>

        {devices && devices.length === 0 ? (
          <div className="text-center py-12">
            <p className="text-gray-500 dark:text-gray-400">No devices found</p>
          </div>
        ) : (
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
            {devices?.map((device) => (
              <Link
                key={device.id}
                to={`/device/${device.id}`}
                className="bg-white dark:bg-gray-800 rounded-lg shadow-lg p-6 border border-gray-200 dark:border-gray-700 hover:shadow-xl transition-shadow"
              >
                <div className="flex items-start justify-between mb-4">
                  <div>
                    <h2 className="text-xl font-semibold text-gray-900 dark:text-white">
                      {device.name}
                    </h2>
                    <p className="text-sm text-gray-500 dark:text-gray-400">
                      {device.type}
                    </p>
                  </div>
                  <DeviceStatus status={device.status} />
                </div>

                <div className="text-sm text-gray-500 dark:text-gray-400">
                  Last seen: {formatTimestamp(device.lastSeen)}
                </div>
              </Link>
            ))}
          </div>
        )}
      </div>
    </div>
  );
}
