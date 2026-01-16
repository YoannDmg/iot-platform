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
import { Button } from "@/components/ui/button"
import { IconArrowLeft } from "@tabler/icons-react"
import { AddDeviceForm, type AddDeviceFormValues } from "../components/add-device-form"

export function AddDevicePage() {
  const navigate = useNavigate()

  const handleSubmit = (data: AddDeviceFormValues) => {
    // TODO: Call API to create device
    console.log("Creating device:", data)
    navigate("/devices")
  }

  const handleCancel = () => {
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
          <Button variant="ghost" onClick={handleCancel}>
            <IconArrowLeft className="mr-2 h-4 w-4" />
            Back
          </Button>
        </div>

        <div className="max-w-2xl">
          <AddDeviceForm onSubmit={handleSubmit} onCancel={handleCancel} />
        </div>
      </div>
    </>
  )
}
