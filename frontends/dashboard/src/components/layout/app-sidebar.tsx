import { Link, useLocation } from 'react-router-dom';
import {
  LayoutDashboard,
  Cpu,
  Settings,
  Users,
  Activity,
} from 'lucide-react';
import {
  Sidebar,
  SidebarContent,
  SidebarFooter,
  SidebarGroup,
  SidebarGroupContent,
  SidebarGroupLabel,
  SidebarHeader,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
} from '@/components/ui/sidebar';
import { useSidebar } from '@/lib/sidebar-context';
import { Separator } from '@/components/ui/separator';

const navigationItems = [
  {
    title: 'Dashboard',
    icon: LayoutDashboard,
    href: '/dashboard',
  },
  {
    title: 'Devices',
    icon: Cpu,
    href: '/devices',
  },
  {
    title: 'Activity',
    icon: Activity,
    href: '/activity',
  },
  {
    title: 'Users',
    icon: Users,
    href: '/users',
  },
  {
    title: 'Settings',
    icon: Settings,
    href: '/settings',
  },
];

export function AppSidebar() {
  const location = useLocation();
  const { isOpen } = useSidebar();

  return (
    <Sidebar>
      <SidebarHeader>
        <div className="flex items-center gap-3">
          <div className="flex h-9 w-9 shrink-0 items-center justify-center rounded-lg bg-primary text-primary-foreground">
            <Activity className="h-5 w-5" />
          </div>
          {isOpen && (
            <div className="flex flex-col">
              <span className="text-sm font-semibold">IoT Platform</span>
              <span className="text-xs text-muted-foreground">Device Manager</span>
            </div>
          )}
        </div>
      </SidebarHeader>


      <SidebarContent>
        <SidebarGroup>
          <SidebarGroupLabel>Navigation</SidebarGroupLabel>
          <SidebarGroupContent>
            <SidebarMenu>
              {navigationItems.map((item) => {
                const isActive = location.pathname === item.href ||
                                location.pathname.startsWith(item.href + '/');
                return (
                  <SidebarMenuItem key={item.href}>
                    <SidebarMenuButton asChild isActive={isActive}>
                      <Link to={item.href}>
                        <item.icon className="h-4 w-4 shrink-0" />
                        {isOpen && <span>{item.title}</span>}
                      </Link>
                    </SidebarMenuButton>
                  </SidebarMenuItem>
                );
              })}
            </SidebarMenu>
          </SidebarGroupContent>
        </SidebarGroup>
      </SidebarContent>

      <Separator className="mx-3" />

      <SidebarFooter>
        {isOpen ? (
          <div className="rounded-lg bg-muted/50 px-3 py-2">
            <p className="text-xs font-medium text-muted-foreground">Version</p>
            <p className="text-sm font-semibold">1.0.0</p>
          </div>
        ) : (
          <div className="text-center text-xs text-muted-foreground">
            v1.0
          </div>
        )}
      </SidebarFooter>
    </Sidebar>
  );
}