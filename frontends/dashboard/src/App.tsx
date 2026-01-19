import { BrowserRouter, Routes, Route } from "react-router-dom"
import { DashboardLayout } from "@/layouts/dashboard-layout"
import { HomePage } from "@/features/home"
import { DevicesPage, DeviceDetailPage, AddDevicePage } from "@/features/devices"
import { LoginPage, ProtectedRoute } from "@/features/auth"

export function App() {
  return (
    <BrowserRouter>
      <Routes>
        <Route path="/login" element={<LoginPage />} />
        <Route element={<ProtectedRoute />}>
          <Route element={<DashboardLayout />}>
            <Route path="/" element={<HomePage />} />
            <Route path="/devices" element={<DevicesPage />} />
            <Route path="/devices/new" element={<AddDevicePage />} />
            <Route path="/devices/:id" element={<DeviceDetailPage />} />
          </Route>
        </Route>
      </Routes>
    </BrowserRouter>
  )
}

export default App
