import { BrowserRouter, Routes, Route } from 'react-router-dom';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { DeviceList } from './pages/DeviceList';
import { DeviceDetail } from './pages/DeviceDetail';

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      retry: 1,
      refetchOnWindowFocus: false,
    },
  },
});

function App() {
  return (
    <QueryClientProvider client={queryClient}>
      <BrowserRouter>
        <Routes>
          <Route path="/" element={<DeviceList />} />
          <Route path="/device/:id" element={<DeviceDetail />} />
        </Routes>
      </BrowserRouter>
    </QueryClientProvider>
  );
}

export default App;
