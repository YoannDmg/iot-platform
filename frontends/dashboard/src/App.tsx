import { DashboardLayout } from "@/layouts/dashboard-layout"
import { HomePage } from "@/features/home"

export function App() {
  return (
    <DashboardLayout>
      <HomePage />
    </DashboardLayout>
  )
}

export default App