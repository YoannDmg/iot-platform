import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Badge } from "@/components/ui/badge"
import { cn } from "@/lib/utils"
import { IconAlertTriangle } from "@tabler/icons-react"

interface Alert {
  id: string
  deviceName: string
  message: string
  type: "warning" | "error" | "info"
  timestamp: string
}

interface ActiveAlertsProps {
  alerts: Alert[]
  className?: string
}

const alertConfig: Record<Alert["type"], { variant: "destructive" | "secondary" | "outline"; icon: string }> = {
  error: { variant: "destructive", icon: "text-red-500" },
  warning: { variant: "secondary", icon: "text-yellow-500" },
  info: { variant: "outline", icon: "text-blue-500" },
}

export function ActiveAlerts({ alerts, className }: ActiveAlertsProps) {
  return (
    <Card className={cn("", className)}>
      <CardHeader className="flex flex-row items-center gap-2">
        <IconAlertTriangle className="h-5 w-5 text-yellow-500" />
        <CardTitle>Active Alerts</CardTitle>
        {alerts.length > 0 && (
          <Badge variant="destructive" className="ml-auto">
            {alerts.length}
          </Badge>
        )}
      </CardHeader>
      <CardContent>
        {alerts.length === 0 ? (
          <p className="text-muted-foreground text-sm text-center py-4">
            No active alerts
          </p>
        ) : (
          <div className="space-y-3">
            {alerts.map((alert) => {
              const config = alertConfig[alert.type]
              return (
                <div
                  key={alert.id}
                  className="flex items-start gap-3 rounded-lg border p-3"
                >
                  <div className={cn("mt-0.5", config.icon)}>
                    <IconAlertTriangle className="h-4 w-4" />
                  </div>
                  <div className="flex-1 min-w-0">
                    <div className="flex items-center gap-2">
                      <span className="font-medium text-sm">
                        {alert.deviceName}
                      </span>
                      <Badge variant={config.variant} className="text-[0.6rem]">
                        {alert.type}
                      </Badge>
                    </div>
                    <p className="text-muted-foreground text-xs mt-0.5">
                      {alert.message}
                    </p>
                  </div>
                  <span className="text-muted-foreground text-xs whitespace-nowrap">
                    {alert.timestamp}
                  </span>
                </div>
              )
            })}
          </div>
        )}
      </CardContent>
    </Card>
  )
}
