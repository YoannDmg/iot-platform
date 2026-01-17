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
import { useDevices, useDeleteDevice } from "@/hooks/use-devices"
import type { Device } from "@/types/device"
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from "@/components/ui/alert-dialog"

export function DevicesPage() {
  const navigate = useNavigate()
  const [search, setSearch] = useState("")
  const [deviceToDelete, setDeviceToDelete] = useState<Device | null>(null)

  const { devices, loading, error } = useDevices()
  const { deleteDevice, loading: deleteLoading } = useDeleteDevice()

  const filteredDevices = useMemo(() => {
    if (!search) return devices
    const searchLower = search.toLowerCase()
    return devices.filter(
      (device) =>
        device.name.toLowerCase().includes(searchLower) ||
        device.type.toLowerCase().includes(searchLower)
    )
  }, [search, devices])

  const handleView = (device: Device) => {
    navigate(`/devices/${device.id}`)
  }

  const handleEdit = (device: Device) => {
    navigate(`/devices/${device.id}/edit`)
  }

  const handleDelete = (device: Device) => {
    setDeviceToDelete(device)
  }

  const confirmDelete = async () => {
    if (deviceToDelete) {
      await deleteDevice(deviceToDelete.id)
      setDeviceToDelete(null)
    }
  }

  if (error) {
    return (
      <div className="flex flex-1 items-center justify-center">
        <p className="text-destructive">Error loading devices: {error.message}</p>
      </div>
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
            {loading ? (
              <div className="flex h-32 items-center justify-center">
                <p className="text-muted-foreground">Loading devices...</p>
              </div>
            ) : (
              <DeviceTable
                devices={filteredDevices}
                onView={handleView}
                onEdit={handleEdit}
                onDelete={handleDelete}
              />
            )}
          </CardContent>
        </Card>
      </div>

      <AlertDialog open={!!deviceToDelete} onOpenChange={() => setDeviceToDelete(null)}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Delete Device</AlertDialogTitle>
            <AlertDialogDescription>
              Are you sure you want to delete "{deviceToDelete?.name}"? This action cannot be undone.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>Cancel</AlertDialogCancel>
            <AlertDialogAction
              onClick={confirmDelete}
              disabled={deleteLoading}
              className="bg-destructive text-destructive-foreground hover:bg-destructive/90"
            >
              {deleteLoading ? "Deleting..." : "Delete"}
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </>
  )
}
