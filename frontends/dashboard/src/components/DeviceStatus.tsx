import { Badge } from '@/components/ui/badge';

interface DeviceStatusProps {
  status: 'UNKNOWN' | 'ONLINE' | 'OFFLINE' | 'ERROR' | 'MAINTENANCE';
}

export function DeviceStatus({ status }: DeviceStatusProps) {
  const statusConfig = {
    UNKNOWN: { color: 'bg-gray-400', variant: 'secondary' as const },
    ONLINE: { color: 'bg-green-500', variant: 'default' as const },
    OFFLINE: { color: 'bg-gray-500', variant: 'secondary' as const },
    ERROR: { color: 'bg-red-500', variant: 'destructive' as const },
    MAINTENANCE: { color: 'bg-blue-500', variant: 'default' as const },
  };

  const config = statusConfig[status];

  return (
    <Badge variant={config.variant} className="gap-1.5">
      <div className={`w-2 h-2 rounded-full ${config.color}`} />
      {status}
    </Badge>
  );
}
