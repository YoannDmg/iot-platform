import { useParams, useNavigate } from "react-router-dom"
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
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Button } from "@/components/ui/button"
import { Badge } from "@/components/ui/badge"
import { IconArrowLeft, IconEdit } from "@tabler/icons-react"
import { DeviceStatusBadge } from "../components"
import type { Device } from "@/shared/types"

// Mock data - à remplacer par les vraies données de l'API
const mockDevices: Record<string, Device> = {
  "1": {
    id: "1",
    name: "Temperature Sensor #1",
    type: "Temperature",
    status: "ONLINE",
    createdAt: 1704067200,
    lastSeen: Math.floor(Date.now() / 1000) - 120,
    metadata: [
      { key: "location", value: "Room A" },
      { key: "model", value: "TMP-100" },
      { key: "firmware", value: "v2.1.0" },
    ],
  },
  "2": {
    id: "2",
    name: "Humidity Sensor #2",
    type: "Humidity",
    status: "OFFLINE",
    createdAt: 1704067200,
    lastSeen: Math.floor(Date.now() / 1000) - 7200,
    metadata: [
      { key: "location", value: "Room B" },
      { key: "model", value: "HUM-200" },
    ],
  },
  "3": {
    id: "3",
    name: "Motion Detector #3",
    type: "Motion",
    status: "ONLINE",
    createdAt: 1704067200,
    lastSeen: Math.floor(Date.now() / 1000) - 30,
    metadata: [{ key: "location", value: "Entrance" }],
  },
  "4": {
    id: "4",
    name: "Light Controller #4",
    type: "Light",
    status: "MAINTENANCE",
    createdAt: 1704067200,
    lastSeen: Math.floor(Date.now() / 1000) - 86400,
    metadata: [{ key: "location", value: "Living Room" }],
  },
  "5": {
    id: "5",
    name: "Door Sensor #5",
    type: "Door",
    status: "ONLINE",
    createdAt: 1704067200,
    lastSeen: Math.floor(Date.now() / 1000) - 60,
    metadata: [{ key: "location", value: "Front Door" }],
  },
  "6": {
    id: "6",
    name: "Smoke Detector #6",
    type: "Smoke",
    status: "ERROR",
    createdAt: 1704067200,
    lastSeen: Math.floor(Date.now() / 1000) - 3600,
    metadata: [{ key: "location", value: "Kitchen" }],
  },
}

function formatDate(timestamp: number): string {
  return new Date(timestamp * 1000).toLocaleDateString("en-US", {
    year: "numeric",
    month: "long",
    day: "numeric",
    hour: "2-digit",
    minute: "2-digit",
  })
}

export function DeviceDetailPage() {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()

  const device = id ? mockDevices[id] : null

  if (!device) {
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
                  <BreadcrumbLink href="/">Home</BreadcrumbLink>
                </BreadcrumbItem>
                <BreadcrumbSeparator className="hidden md:block" />
                <BreadcrumbItem>
                  <BreadcrumbLink href="/devices">Devices</BreadcrumbLink>
                </BreadcrumbItem>
                <BreadcrumbSeparator className="hidden md:block" />
                <BreadcrumbItem>
                  <BreadcrumbPage>Not Found</BreadcrumbPage>
                </BreadcrumbItem>
              </BreadcrumbList>
            </Breadcrumb>
          </div>
        </header>
        <div className="flex flex-1 flex-col items-center justify-center gap-4 p-4">
          <p className="text-muted-foreground">Device not found</p>
          <Button variant="outline" onClick={() => navigate("/devices")}>
            <IconArrowLeft className="mr-2 h-4 w-4" />
            Back to devices
          </Button>
        </div>
      </>
    )
  }

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
                <BreadcrumbLink href="/">Home</BreadcrumbLink>
              </BreadcrumbItem>
              <BreadcrumbSeparator className="hidden md:block" />
              <BreadcrumbItem className="hidden md:block">
                <BreadcrumbLink href="/devices">Devices</BreadcrumbLink>
              </BreadcrumbItem>
              <BreadcrumbSeparator className="hidden md:block" />
              <BreadcrumbItem>
                <BreadcrumbPage>{device.name}</BreadcrumbPage>
              </BreadcrumbItem>
            </BreadcrumbList>
          </Breadcrumb>
        </div>
      </header>

      <div className="flex flex-1 flex-col gap-4 p-4 pt-0">
        {/* Header with back button */}
        <div className="flex items-center justify-between">
          <Button variant="ghost" onClick={() => navigate("/devices")}>
            <IconArrowLeft className="mr-2 h-4 w-4" />
            Back
          </Button>
          <Button variant="outline">
            <IconEdit className="mr-2 h-4 w-4" />
            Edit
          </Button>
        </div>

        {/* Device Info */}
        <div className="grid gap-4 md:grid-cols-2">
          <Card>
            <CardHeader>
              <CardTitle className="flex items-center justify-between">
                {device.name}
                <DeviceStatusBadge status={device.status} />
              </CardTitle>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="grid grid-cols-2 gap-4">
                <div>
                  <p className="text-muted-foreground text-sm">Type</p>
                  <Badge variant="secondary">{device.type}</Badge>
                </div>
                <div>
                  <p className="text-muted-foreground text-sm">ID</p>
                  <p className="font-mono text-sm">{device.id}</p>
                </div>
                <div>
                  <p className="text-muted-foreground text-sm">Created</p>
                  <p className="text-sm">{formatDate(device.createdAt)}</p>
                </div>
                <div>
                  <p className="text-muted-foreground text-sm">Last Seen</p>
                  <p className="text-sm">{formatDate(device.lastSeen)}</p>
                </div>
              </div>
            </CardContent>
          </Card>

          <Card>
            <CardHeader>
              <CardTitle>Metadata</CardTitle>
            </CardHeader>
            <CardContent>
              {device.metadata.length === 0 ? (
                <p className="text-muted-foreground text-sm">No metadata</p>
              ) : (
                <div className="space-y-2">
                  {device.metadata.map((meta) => (
                    <div
                      key={meta.key}
                      className="flex items-center justify-between rounded-lg border p-2"
                    >
                      <span className="text-muted-foreground text-sm">
                        {meta.key}
                      </span>
                      <span className="text-sm font-medium">{meta.value}</span>
                    </div>
                  ))}
                </div>
              )}
            </CardContent>
          </Card>
        </div>

        {/* Placeholder for future sections */}
        <Card>
          <CardHeader>
            <CardTitle>Activity Log</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="flex h-[200px] items-center justify-center rounded-lg border border-dashed">
              <span className="text-muted-foreground text-sm">
                Activity log coming soon...
              </span>
            </div>
          </CardContent>
        </Card>
      </div>
    </>
  )
}
