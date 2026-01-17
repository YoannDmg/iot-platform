import {
  Breadcrumb,
  BreadcrumbItem,
  BreadcrumbLink,
  BreadcrumbList,
  BreadcrumbPage,
  BreadcrumbSeparator,
} from "@/components/ui/breadcrumb"
import { Separator } from "@/components/ui/separator"
import { SidebarTrigger } from "@/components/ui/sidebar"
import {
  IconDevices,
  IconWifi,
  IconWifiOff,
  IconAlertTriangle,
} from "@tabler/icons-react"
import { StatsCard, RecentActivity, ChartCard, ActiveAlerts } from "../components"
import { useStats, useDevices } from "@/hooks/use-devices"

// Mock data for activities and alerts (not yet available in API)
const mockActivities = [
  {
    id: "1",
    deviceId: "23",
    deviceName: "Device #23",
    deviceType: "Temperature sensor",
    status: "ONLINE" as const,
    action: "Connected successfully",
    timestamp: "2s ago",
  },
  {
    id: "2",
    deviceId: "15",
    deviceName: "Device #15",
    deviceType: "Humidity sensor",
    status: "OFFLINE" as const,
    action: "Connection lost",
    timestamp: "5m ago",
  },
  {
    id: "3",
    deviceId: "08",
    deviceName: "Device #08",
    deviceType: "Motion detector",
    status: "MAINTENANCE" as const,
    action: "Entering maintenance mode",
    timestamp: "12m ago",
  },
  {
    id: "4",
    deviceId: "42",
    deviceName: "Device #42",
    deviceType: "Light controller",
    status: "ONLINE" as const,
    action: "Data received: brightness=75%",
    timestamp: "15m ago",
  },
]

const mockAlerts = [
  {
    id: "1",
    deviceName: "Device #15",
    message: "Battery low (5%)",
    type: "warning" as const,
    timestamp: "5m ago",
  },
  {
    id: "2",
    deviceName: "Device #23",
    message: "Temperature threshold exceeded (35°C)",
    type: "error" as const,
    timestamp: "10m ago",
  },
  {
    id: "3",
    deviceName: "Device #08",
    message: "No data received for 30 minutes",
    type: "warning" as const,
    timestamp: "30m ago",
  },
]

export function HomePage() {
  const { stats, loading: statsLoading } = useStats()
  const { devices } = useDevices({ status: "ERROR" })

  const alertsCount = devices.length + mockAlerts.length

  const totalDevices = stats?.totalDevices ?? 0
  const onlineDevices = stats?.onlineDevices ?? 0
  const offlineDevices = stats?.offlineDevices ?? 0

  const onlinePercentage = totalDevices > 0
    ? Math.round((onlineDevices / totalDevices) * 100)
    : 0
  const offlinePercentage = totalDevices > 0
    ? Math.round((offlineDevices / totalDevices) * 100)
    : 0

  return (
    <>
      <header className="flex h-16 shrink-0 items-center gap-2 transition-[width,height] ease-linear group-has-data-[collapsible=icon]/sidebar-wrapper:h-12">
        <div className="flex items-center gap-2 px-4">
          <SidebarTrigger className="-ml-1" />
          <Separator
            orientation="vertical"
            className="mr-2 data-[orientation=vertical]:h-4"
          />
          <Breadcrumb>
            <BreadcrumbList>
              <BreadcrumbItem className="hidden md:block">
                <BreadcrumbLink href="#">Home</BreadcrumbLink>
              </BreadcrumbItem>
              <BreadcrumbSeparator className="hidden md:block" />
              <BreadcrumbItem>
                <BreadcrumbPage>Overview</BreadcrumbPage>
              </BreadcrumbItem>
            </BreadcrumbList>
          </Breadcrumb>
        </div>
      </header>

      <div className="flex flex-1 flex-col gap-4 p-4 pt-0">
        {/* Section 1: KPIs */}
        <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
          <StatsCard
            title="Total Devices"
            value={statsLoading ? "-" : totalDevices}
            icon={<IconDevices className="h-5 w-5" />}
          />
          <StatsCard
            title="Online Devices"
            value={statsLoading ? "-" : onlineDevices}
            subtitle={statsLoading ? "" : `${onlinePercentage}%`}
            icon={<IconWifi className="h-5 w-5" />}
          />
          <StatsCard
            title="Offline Devices"
            value={statsLoading ? "-" : offlineDevices}
            subtitle={statsLoading ? "" : `${offlinePercentage}%`}
            icon={<IconWifiOff className="h-5 w-5" />}
          />
          <StatsCard
            title="Active Alerts"
            value={statsLoading ? "-" : alertsCount}
            icon={<IconAlertTriangle className="h-5 w-5" />}
            className={alertsCount > 0 ? "border-yellow-500/50" : ""}
          />
        </div>

        {/* Section 2: Recent Activity */}
        <RecentActivity activities={mockActivities} />

        {/* Section 3: Charts */}
        <div className="grid gap-4 md:grid-cols-2">
          <ChartCard title="Device Activity (24h)" />
          <ChartCard title="Data Volume (24h)" />
        </div>

        {/* Section 4: Active Alerts */}
        <ActiveAlerts alerts={mockAlerts} />
      </div>
    </>
  )
}
