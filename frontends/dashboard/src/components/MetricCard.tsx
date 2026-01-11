import { Card } from '@/components/ui/card';

interface MetricCardProps {
  label: string;
  value: string | number;
  unit?: string;
  icon?: string;
}

export function MetricCard({ label, value, unit, icon }: MetricCardProps) {
  return (
    <Card className="p-6">
      <div className="flex items-center justify-between mb-2">
        <span className="text-sm text-muted-foreground">{label}</span>
        {icon && <span className="text-xl">{icon}</span>}
      </div>
      <div className="text-2xl font-bold">
        {value}
        {unit && <span className="text-lg text-muted-foreground ml-1">{unit}</span>}
      </div>
    </Card>
  );
}
