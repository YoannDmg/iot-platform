interface MetricCardProps {
  label: string;
  value: string | number;
  unit?: string;
  icon?: string;
}

export function MetricCard({ label, value, unit, icon }: MetricCardProps) {
  return (
    <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-6 border border-gray-200 dark:border-gray-700">
      <div className="flex items-center justify-between mb-2">
        <span className="text-sm text-gray-600 dark:text-gray-400">{label}</span>
        {icon && <span className="text-xl">{icon}</span>}
      </div>
      <div className="text-2xl font-bold text-gray-900 dark:text-white">
        {value}
        {unit && <span className="text-lg text-gray-500 ml-1">{unit}</span>}
      </div>
    </div>
  );
}
