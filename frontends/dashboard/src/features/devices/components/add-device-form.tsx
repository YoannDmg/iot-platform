import { useForm } from "@tanstack/react-form"
import { z } from "zod"
import { Card, CardContent, CardFooter, CardHeader, CardTitle, CardDescription } from "@/components/ui/card"
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
import { Separator } from "@/components/ui/separator"
import { IconPlus, IconTrash } from "@tabler/icons-react"

const deviceTypes = [
  { value: "temperature", label: "Temperature Sensor" },
  { value: "humidity", label: "Humidity Sensor" },
  { value: "motion", label: "Motion Detector" },
  { value: "light", label: "Light Controller" },
  { value: "door", label: "Door Sensor" },
  { value: "smoke", label: "Smoke Detector" },
  { value: "camera", label: "Camera" },
  { value: "other", label: "Other" },
] as const

const metadataEntrySchema = z.object({
  key: z.string(),
  value: z.string(),
})

const addDeviceSchema = z.object({
  name: z.string().min(1, "Device name is required"),
  type: z.string().min(1, "Device type is required"),
  metadata: z.array(metadataEntrySchema),
})

export type AddDeviceFormValues = z.infer<typeof addDeviceSchema>

export interface AddDeviceFormProps {
  onSubmit: (data: AddDeviceFormValues) => void
  onCancel: () => void
}

export function AddDeviceForm({ onSubmit, onCancel }: AddDeviceFormProps) {
  const form = useForm({
    defaultValues: {
      name: "",
      type: "",
      metadata: [{ key: "", value: "" }],
    } as AddDeviceFormValues,
    validators: {
      onSubmit: addDeviceSchema,
    },
    onSubmit: async ({ value }) => {
      const filteredMetadata = value.metadata.filter(
        (entry) => entry.key.trim() && entry.value.trim()
      )
      onSubmit({
        ...value,
        name: value.name.trim(),
        metadata: filteredMetadata,
      })
    },
  })

  return (
    <form
      onSubmit={(e) => {
        e.preventDefault()
        e.stopPropagation()
        form.handleSubmit()
      }}
    >
      <Card>
        <CardHeader>
          <CardTitle>Add New Device</CardTitle>
          <CardDescription>
            Enter the information for your new device.
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-6">
          <FieldGroup>
            <form.Field
              name="name"
              children={(field) => {
                const hasError = field.state.meta.isTouched && field.state.meta.errors.length > 0
                return (
                  <Field data-invalid={hasError}>
                    <FieldLabel htmlFor="name">Device Name</FieldLabel>
                    <Input
                      id="name"
                      placeholder="e.g., Living Room Sensor"
                      value={field.state.value}
                      onChange={(e) => field.handleChange(e.target.value)}
                      onBlur={field.handleBlur}
                      aria-invalid={hasError}
                    />
                    {hasError && (
                      <FieldError>
                        {field.state.meta.errors.map((e) => e?.message).filter(Boolean).join(", ")}
                      </FieldError>
                    )}
                  </Field>
                )
              }}
            />

            <form.Field
              name="type"
              children={(field) => {
                const hasError = field.state.meta.isTouched && field.state.meta.errors.length > 0
                return (
                  <Field data-invalid={hasError}>
                    <FieldLabel htmlFor="type">Device Type</FieldLabel>
                    <Select
                      value={field.state.value}
                      onValueChange={(value) => field.handleChange(value ?? "")}
                    >
                      <SelectTrigger className="w-full" aria-invalid={hasError}>
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
                    {hasError && (
                      <FieldError>
                        {field.state.meta.errors.map((e) => e?.message).filter(Boolean).join(", ")}
                      </FieldError>
                    )}
                  </Field>
                )
              }}
            />
          </FieldGroup>

          <Separator />

          <div className="space-y-4">
            <div>
              <h3 className="text-sm font-medium">Metadata</h3>
              <p className="text-sm text-muted-foreground">
                Add custom key-value pairs to store additional information.
              </p>
            </div>
            <FieldGroup>
              <form.Field
                name="metadata"
                mode="array"
                children={(field) => (
                  <>
                    {field.state.value.map((_, index) => (
                      <div key={index} className="flex items-end gap-2">
                        <form.Field
                          name={`metadata[${index}].key`}
                          children={(keyField) => (
                            <Field className="flex-1">
                              <FieldLabel htmlFor={`meta-key-${index}`}>
                                {index === 0 ? "Key" : <span className="sr-only">Key</span>}
                              </FieldLabel>
                              <Input
                                id={`meta-key-${index}`}
                                placeholder="e.g., location"
                                value={keyField.state.value}
                                onChange={(e) => keyField.handleChange(e.target.value)}
                              />
                            </Field>
                          )}
                        />
                        <form.Field
                          name={`metadata[${index}].value`}
                          children={(valueField) => (
                            <Field className="flex-1">
                              <FieldLabel htmlFor={`meta-value-${index}`}>
                                {index === 0 ? "Value" : <span className="sr-only">Value</span>}
                              </FieldLabel>
                              <Input
                                id={`meta-value-${index}`}
                                placeholder="e.g., Room A"
                                value={valueField.state.value}
                                onChange={(e) => valueField.handleChange(e.target.value)}
                              />
                            </Field>
                          )}
                        />
                        <Button
                          type="button"
                          variant="ghost"
                          size="icon"
                          onClick={() => field.removeValue(index)}
                          disabled={field.state.value.length === 1}
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
                      onClick={() => field.pushValue({ key: "", value: "" })}
                    >
                      <IconPlus className="mr-2 h-4 w-4" />
                      Add Metadata
                    </Button>
                    <FieldDescription>
                      Metadata is optional. You can add information like location,
                      firmware version, or any custom attributes.
                    </FieldDescription>
                  </>
                )}
              />
            </FieldGroup>
          </div>
        </CardContent>
        <CardFooter className="flex justify-end gap-2">
          <Button type="button" variant="outline" onClick={onCancel}>
            Cancel
          </Button>
          <Button type="submit">
            <IconPlus className="mr-2 h-4 w-4" />
            Create Device
          </Button>
        </CardFooter>
      </Card>
    </form>
  )
}
