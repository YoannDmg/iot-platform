import { BrowserRouter, Routes, Route } from "react-router-dom"
import { DashboardLayout } from "@/layouts/dashboard-layout"
import { HomePage } from "@/features/home"
import { DevicesPage, DeviceDetailPage } from "@/features/devices"

export function App() {
  return (
    <BrowserRouter>
      <DashboardLayout>
        <Routes>
          <Route path="/" element={<HomePage />} />
          <Route path="/devices" element={<DevicesPage />} />
          <Route path="/devices/:id" element={<DeviceDetailPage />} />
        </Routes>
      </DashboardLayout>
    </BrowserRouter>
  )
}

export default App
