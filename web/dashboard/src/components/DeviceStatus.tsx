interface DeviceStatusProps {
  status: 'UNKNOWN' | 'ONLINE' | 'OFFLINE' | 'ERROR' | 'MAINTENANCE';
}

export function DeviceStatus({ status }: DeviceStatusProps) {
  const statusColors = {
    UNKNOWN: 'bg-gray-400',
    ONLINE: 'bg-green-500',
    OFFLINE: 'bg-red-500',
    ERROR: 'bg-orange-500',
    MAINTENANCE: 'bg-blue-500',
  };

  return (
    <div className="flex items-center gap-2">
      <div className={`w-3 h-3 rounded-full ${statusColors[status]}`} />
      <span className="text-sm font-medium">{status}</span>
    </div>
  );
}
