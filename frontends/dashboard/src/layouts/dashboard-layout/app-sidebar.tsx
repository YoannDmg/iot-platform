import {
  IconHome,
  IconDevices,
  IconSettings,
  IconBook,
} from "@tabler/icons-react"

import { NavMain } from "./nav-main"
import { SidebarLogo } from "./sidebar-logo"
import { ThemeToggle } from "./theme-toggle"
import {
  Sidebar,
  SidebarContent,
  SidebarHeader,
  SidebarRail,
} from "@/components/ui/sidebar"

const data = {
  navMain: [
    {
      title: "Home",
      url: "/",
      icon: IconHome,
      isActive: true,
      items: [
        { title: "Overview", url: "/" },
        { title: "Analytics", url: "#" },
      ],
    },
    {
      title: "Devices",
      url: "/devices",
      icon: IconDevices,
      items: [
        { title: "All Devices", url: "/devices" },
        { title: "Groups", url: "#" },
        { title: "Add Device", url: "#" },
      ],
    },
    {
      title: "Documentation",
      url: "#",
      icon: IconBook,
      items: [
        { title: "Introduction", url: "#" },
        { title: "Get Started", url: "#" },
      ],
    },
    {
      title: "Settings",
      url: "#",
      icon: IconSettings,
      items: [
        { title: "General", url: "#" },
        { title: "Team", url: "#" },
        { title: "Billing", url: "#" },
      ],
    },
  ],
}

export function AppSidebar({ ...props }: React.ComponentProps<typeof Sidebar>) {
  return (
    <Sidebar collapsible="icon" {...props}>
      <SidebarHeader>
        <SidebarLogo />
      </SidebarHeader>
      <SidebarContent>
        <NavMain items={data.navMain} />
        <ThemeToggle className="mt-auto" />
      </SidebarContent>
      <SidebarRail />
    </Sidebar>
  )
}