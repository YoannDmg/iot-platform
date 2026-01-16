import { useState } from "react"
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
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select"
import {
  Field,
  FieldLabel,
  FieldDescription,
  FieldError,
  FieldGroup,
} from "@/components/ui/field"
import { IconArrowLeft, IconPlus, IconTrash } from "@tabler/icons-react"

const deviceTypes = [
  { value: "temperature", label: "Temperature Sensor" },
  { value: "humidity", label: "Humidity Sensor" },
  { value: "motion", label: "Motion Detector" },
  { value: "light", label: "Light Controller" },
  { value: "door", label: "Door Sensor" },
  { value: "smoke", label: "Smoke Detector" },
  { value: "camera", label: "Camera" },
  { value: "other", label: "Other" },
]

interface MetadataEntry {
  id: string
  key: string
  value: string
}

export function AddDevicePage() {
  const navigate = useNavigate()
  const [name, setName] = useState("")
  const [type, setType] = useState("")
  const [metadata, setMetadata] = useState<MetadataEntry[]>([
    { id: crypto.randomUUID(), key: "", value: "" },
  ])
  const [errors, setErrors] = useState<Record<string, string>>({})

  const addMetadataEntry = () => {
    setMetadata([...metadata, { id: crypto.randomUUID(), key: "", value: "" }])
  }

  const removeMetadataEntry = (id: string) => {
    if (metadata.length > 1) {
      setMetadata(metadata.filter((entry) => entry.id !== id))
    }
  }

  const updateMetadataEntry = (
    id: string,
    field: "key" | "value",
    value: string
  ) => {
    setMetadata(
      metadata.map((entry) =>
        entry.id === id ? { ...entry, [field]: value } : entry
      )
    )
  }

  const validateForm = (): boolean => {
    const newErrors: Record<string, string> = {}

    if (!name.trim()) {
      newErrors.name = "Device name is required"
    }

    if (!type) {
      newErrors.type = "Device type is required"
    }

    setErrors(newErrors)
    return Object.keys(newErrors).length === 0
  }

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()

    if (!validateForm()) {
      return
    }

    // Filter out empty metadata entries
    const filteredMetadata = metadata.filter(
      (entry) => entry.key.trim() && entry.value.trim()
    )

    const deviceData = {
      name: name.trim(),
      type,
      metadata: filteredMetadata.map(({ key, value }) => ({ key, value })),
    }

    // TODO: Call API to create device
    console.log("Creating device:", deviceData)

    // Navigate back to devices list
    navigate("/devices")
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
                <BreadcrumbPage>Add Device</BreadcrumbPage>
              </BreadcrumbItem>
            </BreadcrumbList>
          </Breadcrumb>
        </div>
      </header>

      <div className="flex flex-1 flex-col gap-4 p-4 pt-0">
        <div className="flex items-center">
          <Button variant="ghost" onClick={() => navigate("/devices")}>
            <IconArrowLeft className="mr-2 h-4 w-4" />
            Back
          </Button>
        </div>

        <form onSubmit={handleSubmit}>
          <div className="grid gap-4 md:grid-cols-2">
            {/* Basic Information */}
            <Card>
              <CardHeader>
                <CardTitle>Device Information</CardTitle>
                <CardDescription>
                  Enter the basic information for your new device.
                </CardDescription>
              </CardHeader>
              <CardContent>
                <FieldGroup>
                  <Field data-invalid={!!errors.name}>
                    <FieldLabel htmlFor="name">Device Name</FieldLabel>
                    <Input
                      id="name"
                      placeholder="e.g., Living Room Sensor"
                      value={name}
                      onChange={(e) => setName(e.target.value)}
                      aria-invalid={!!errors.name}
                    />
                    {errors.name && <FieldError>{errors.name}</FieldError>}
                  </Field>

                  <Field data-invalid={!!errors.type}>
                    <FieldLabel htmlFor="type">Device Type</FieldLabel>
                    <Select value={type} onValueChange={(value) => setType(value ?? "")}>
                      <SelectTrigger className="w-full" aria-invalid={!!errors.type}>
                        <SelectValue placeholder="Select a device type" />
                      </SelectTrigger>
                      <SelectContent>
                        {deviceTypes.map((deviceType) => (
                          <SelectItem key={deviceType.value} value={deviceType.value}>
                            {deviceType.label}
                          </SelectItem>
                        ))}
                      </SelectContent>
                    </Select>
                    {errors.type && <FieldError>{errors.type}</FieldError>}
                  </Field>
                </FieldGroup>
              </CardContent>
            </Card>

            {/* Metadata */}
            <Card>
              <CardHeader>
                <CardTitle>Metadata</CardTitle>
                <CardDescription>
                  Add custom key-value pairs to store additional information.
                </CardDescription>
              </CardHeader>
              <CardContent>
                <FieldGroup>
                  {metadata.map((entry, index) => (
                    <div key={entry.id} className="flex items-end gap-2">
                      <Field className="flex-1">
                        <FieldLabel htmlFor={`meta-key-${entry.id}`}>
                          {index === 0 ? "Key" : <span className="sr-only">Key</span>}
                        </FieldLabel>
                        <Input
                          id={`meta-key-${entry.id}`}
                          placeholder="e.g., location"
                          value={entry.key}
                          onChange={(e) =>
                            updateMetadataEntry(entry.id, "key", e.target.value)
                          }
                        />
                      </Field>
                      <Field className="flex-1">
                        <FieldLabel htmlFor={`meta-value-${entry.id}`}>
                          {index === 0 ? "Value" : <span className="sr-only">Value</span>}
                        </FieldLabel>
                        <Input
                          id={`meta-value-${entry.id}`}
                          placeholder="e.g., Room A"
                          value={entry.value}
                          onChange={(e) =>
                            updateMetadataEntry(entry.id, "value", e.target.value)
                          }
                        />
                      </Field>
                      <Button
                        type="button"
                        variant="ghost"
                        size="icon"
                        onClick={() => removeMetadataEntry(entry.id)}
                        disabled={metadata.length === 1}
                      >
                        <IconTrash className="h-4 w-4" />
                        <span className="sr-only">Remove</span>
                      </Button>
                    </div>
                  ))}
                  <Button
                    type="button"
                    variant="outline"
                    size="sm"
                    onClick={addMetadataEntry}
                  >
                    <IconPlus className="mr-2 h-4 w-4" />
                    Add Metadata
                  </Button>
                  <FieldDescription>
                    Metadata is optional. You can add information like location,
                    firmware version, or any custom attributes.
                  </FieldDescription>
                </FieldGroup>
              </CardContent>
            </Card>
          </div>

          {/* Submit Button */}
          <div className="mt-4 flex justify-end gap-2">
            <Button
              type="button"
              variant="outline"
              onClick={() => navigate("/devices")}
            >
              Cancel
            </Button>
            <Button type="submit">
              <IconPlus className="mr-2 h-4 w-4" />
              Create Device
            </Button>
          </div>
        </form>
      </div>
    </>
  )
}
