import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { cn } from "@/lib/utils"

interface StatsCardProps {
  title: string
  value: number | string
  subtitle?: string
  icon: React.ReactNode
  trend?: {
    value: string
    positive?: boolean
  }
  className?: string
}

export function StatsCard({
  title,
  value,
  subtitle,
  icon,
  trend,
  className,
}: StatsCardProps) {
  return (
    <Card className={cn("relative overflow-hidden", className)}>
      <CardHeader className="flex flex-row items-center justify-between pb-2">
        <CardTitle className="text-muted-foreground text-sm font-medium">
          {title}
        </CardTitle>
        <div className="text-muted-foreground">{icon}</div>
      </CardHeader>
      <CardContent>
        <div className="text-3xl font-bold">{value}</div>
        <div className="flex items-center gap-2 mt-1">
          {trend && (
            <span
              className={cn(
                "text-xs font-medium",
                trend.positive ? "text-green-600 dark:text-green-400" : "text-muted-foreground"
              )}
            >
              {trend.value}
            </span>
          )}
          {subtitle && (
            <span className="text-muted-foreground text-xs">{subtitle}</span>
          )}
        </div>
      </CardContent>
    </Card>
  )
}
