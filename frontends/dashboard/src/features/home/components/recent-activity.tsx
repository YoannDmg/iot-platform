import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Badge } from "@/components/ui/badge"
import { cn } from "@/lib/utils"
import type { ActivityItem } from "../types"

// TODO: Placeholder - à retravailler avec le vrai composant
const statusColors: Record<string, string> = {
  ONLINE: "bg-green-500",
  OFFLINE: "bg-red-500",
  ERROR: "bg-red-500",
  MAINTENANCE: "bg-yellow-500",
  UNKNOWN: "bg-gray-500",
}

interface RecentActivityProps {
  activities: ActivityItem[]
  className?: string
}

export function RecentActivity({ activities, className }: RecentActivityProps) {
  return (
    <Card className={cn("", className)}>
      <CardHeader>
        <CardTitle>Recent Activity</CardTitle>
      </CardHeader>
      <CardContent className="space-y-4">
        {activities.map((activity) => (
          <div
            key={activity.id}
            className="flex items-center gap-3 rounded-lg border p-3"
          >
            <div className={cn("h-2.5 w-2.5 rounded-full", statusColors[activity.status] ?? "bg-gray-500")} />
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
        ))}
      </CardContent>
    </Card>
  )
}
