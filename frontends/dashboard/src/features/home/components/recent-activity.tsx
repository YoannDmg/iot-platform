import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Badge } from "@/components/ui/badge"
import { cn } from "@/lib/utils"
import type { DeviceStatus } from "../types"

interface ActivityItem {
  id: string
  deviceId: string
  deviceName: string
  deviceType: string
  status: DeviceStatus
  action: string
  timestamp: string
}

interface RecentActivityProps {
  activities: ActivityItem[]
  className?: string
}

const statusConfig: Record<DeviceStatus, { color: string; label: string }> = {
  ONLINE: { color: "bg-green-500", label: "Online" },
  OFFLINE: { color: "bg-red-500", label: "Offline" },
  ERROR: { color: "bg-red-500", label: "Error" },
  MAINTENANCE: { color: "bg-yellow-500", label: "Maintenance" },
  UNKNOWN: { color: "bg-gray-500", label: "Unknown" },
}

export function RecentActivity({ activities, className }: RecentActivityProps) {
  return (
    <Card className={cn("", className)}>
      <CardHeader>
        <CardTitle>Recent Activity</CardTitle>
      </CardHeader>
      <CardContent className="space-y-4">
        {activities.map((activity) => {
          const config = statusConfig[activity.status]
          return (
            <div
              key={activity.id}
              className="flex items-center gap-3 rounded-lg border p-3"
            >
              <div className={cn("h-2.5 w-2.5 rounded-full", config.color)} />
              <div className="flex-1 min-w-0">
                <div className="flex items-center gap-2">
                  <span className="font-medium truncate">
                    {activity.deviceName}
                  </span>
                  <Badge variant="secondary" className="text-[0.6rem]">
                    {activity.deviceType}
                  </Badge>
                </div>
                <p className="text-muted-foreground text-xs truncate">
                  {activity.action}
                </p>
              </div>
              <div className="text-muted-foreground text-xs whitespace-nowrap">
                {activity.timestamp}
              </div>
            </div>
          )
        })}
      </CardContent>
    </Card>
  )
}
