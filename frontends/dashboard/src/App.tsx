import { DashboardLayout } from "@/layouts/dashboard-layout"
import { DashboardPage } from "@/features/dashboard"

export function App() {
  return (
    <DashboardLayout>
      <DashboardPage />
    </DashboardLayout>
  )
}

export default App