import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table"
import { Button } from "@/components/ui/button"
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu"
import { IconDots, IconEye, IconEdit, IconTrash } from "@tabler/icons-react"
import type { Device } from "@/types/device"
import { DeviceStatusBadge } from "./device-status-badge"

interface DeviceTableProps {
  devices: Device[]
  onView?: (device: Device) => void
  onEdit?: (device: Device) => void
  onDelete?: (device: Device) => void
}

function formatLastSeen(timestamp: number): string {
  const now = Date.now()
  const diff = now - timestamp * 1000
  const minutes = Math.floor(diff / 60000)
  const hours = Math.floor(diff / 3600000)
  const days = Math.floor(diff / 86400000)

  if (minutes < 1) return "Just now"
  if (minutes < 60) return `${minutes}m ago`
  if (hours < 24) return `${hours}h ago`
  return `${days}d ago`
}

export function DeviceTable({
  devices,
  onView,
  onEdit,
  onDelete,
}: DeviceTableProps) {
  return (
    <Table>
      <TableHeader>
        <TableRow>
          <TableHead>Status</TableHead>
          <TableHead>Name</TableHead>
          <TableHead>Type</TableHead>
          <TableHead>Last Activity</TableHead>
          <TableHead className="w-[70px]">Actions</TableHead>
        </TableRow>
      </TableHeader>
      <TableBody>
        {devices.length === 0 ? (
          <TableRow>
            <TableCell colSpan={5} className="h-24 text-center">
              No devices found.
            </TableCell>
          </TableRow>
        ) : (
          devices.map((device) => (
            <TableRow
              key={device.id}
              className="cursor-pointer"
              onClick={() => onView?.(device)}
            >
              <TableCell>
                <DeviceStatusBadge status={device.status} />
              </TableCell>
              <TableCell className="font-medium">{device.name}</TableCell>
              <TableCell className="text-muted-foreground">
                {device.type}
              </TableCell>
              <TableCell className="text-muted-foreground">
                {formatLastSeen(device.lastSeen)}
              </TableCell>
              <TableCell>
                <DropdownMenu>
                  <DropdownMenuTrigger
                    onClick={(e) => e.stopPropagation()}
                    render={(props) => (
                      <Button variant="ghost" size="icon" {...props}>
                        <IconDots className="h-4 w-4" />
                        <span className="sr-only">Open menu</span>
                      </Button>
                    )}
                  />
                  <DropdownMenuContent align="end">
                    <DropdownMenuItem
                      onClick={(e) => {
                        e.stopPropagation()
                        onView?.(device)
                      }}
                    >
                      <IconEye className="mr-2 h-4 w-4" />
                      View details
                    </DropdownMenuItem>
                    <DropdownMenuItem
                      onClick={(e) => {
                        e.stopPropagation()
                        onEdit?.(device)
                      }}
                    >
                      <IconEdit className="mr-2 h-4 w-4" />
                      Edit
                    </DropdownMenuItem>
                    <DropdownMenuItem
                      onClick={(e) => {
                        e.stopPropagation()
                        onDelete?.(device)
                      }}
                      className="text-destructive focus:text-destructive"
                    >
                      <IconTrash className="mr-2 h-4 w-4" />
                      Delete
                    </DropdownMenuItem>
                  </DropdownMenuContent>
                </DropdownMenu>
              </TableCell>
            </TableRow>
          ))
        )}
      </TableBody>
    </Table>
  )
}
