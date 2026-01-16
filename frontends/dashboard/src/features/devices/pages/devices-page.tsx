import { useState, useMemo } from "react"
import { useNavigate } from "react-router-dom"
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
import { DeviceTable, DeviceSearch } from "../components"
import type { Device } from "@/shared/types"

// Mock data - à remplacer par les vraies données de l'API
const mockDevices: Device[] = [
  {
    id: "1",
    name: "Temperature Sensor #1",
    type: "Temperature",
    status: "ONLINE",
    createdAt: 1704067200,
    lastSeen: Math.floor(Date.now() / 1000) - 120,
    metadata: [{ key: "location", value: "Room A" }],
  },
  {
    id: "2",
    name: "Humidity Sensor #2",
    type: "Humidity",
    status: "OFFLINE",
    createdAt: 1704067200,
    lastSeen: Math.floor(Date.now() / 1000) - 7200,
    metadata: [{ key: "location", value: "Room B" }],
  },
  {
    id: "3",
    name: "Motion Detector #3",
    type: "Motion",
    status: "ONLINE",
    createdAt: 1704067200,
    lastSeen: Math.floor(Date.now() / 1000) - 30,
    metadata: [{ key: "location", value: "Entrance" }],
  },
  {
    id: "4",
    name: "Light Controller #4",
    type: "Light",
    status: "MAINTENANCE",
    createdAt: 1704067200,
    lastSeen: Math.floor(Date.now() / 1000) - 86400,
    metadata: [{ key: "location", value: "Living Room" }],
  },
  {
    id: "5",
    name: "Door Sensor #5",
    type: "Door",
    status: "ONLINE",
    createdAt: 1704067200,
    lastSeen: Math.floor(Date.now() / 1000) - 60,
    metadata: [{ key: "location", value: "Front Door" }],
  },
  {
    id: "6",
    name: "Smoke Detector #6",
    type: "Smoke",
    status: "ERROR",
    createdAt: 1704067200,
    lastSeen: Math.floor(Date.now() / 1000) - 3600,
    metadata: [{ key: "location", value: "Kitchen" }],
  },
]

export function DevicesPage() {
  const navigate = useNavigate()
  const [search, setSearch] = useState("")

  const filteredDevices = useMemo(() => {
    if (!search) return mockDevices
    const searchLower = search.toLowerCase()
    return mockDevices.filter(
      (device) =>
        device.name.toLowerCase().includes(searchLower) ||
        device.type.toLowerCase().includes(searchLower)
    )
  }, [search])

  const handleView = (device: Device) => {
    navigate(`/devices/${device.id}`)
  }

  const handleEdit = (device: Device) => {
    console.log("Edit device:", device)
  }

  const handleDelete = (device: Device) => {
    console.log("Delete device:", device)
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
              <BreadcrumbItem>
                <BreadcrumbPage>Devices</BreadcrumbPage>
              </BreadcrumbItem>
            </BreadcrumbList>
          </Breadcrumb>
        </div>
      </header>

      <div className="flex flex-1 flex-col gap-4 p-4 pt-0">
        <Card>
          <CardHeader className="flex flex-row items-center justify-between">
            <CardTitle>All Devices</CardTitle>
            <div className="w-64">
              <DeviceSearch value={search} onChange={setSearch} />
            </div>
          </CardHeader>
          <CardContent>
            <DeviceTable
              devices={filteredDevices}
              onView={handleView}
              onEdit={handleEdit}
              onDelete={handleDelete}
            />
          </CardContent>
        </Card>
      </div>
    </>
  )
}
