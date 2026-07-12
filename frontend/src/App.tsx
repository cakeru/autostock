import { lazy, Suspense } from 'react'
import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { Toaster, toast } from 'sonner'
import { AuthProvider, useAuth } from '@/contexts/AuthContext'
import { MainLayout } from '@/components/layout/MainLayout'
import { Login } from '@/pages/Login'
import { Dashboard } from '@/pages/Dashboard'
import { Sale } from '@/pages/Sale'
import { Inventory } from '@/pages/Inventory'
import { CustomerDetail } from '@/pages/CustomerDetail'
import { VehicleDetail } from '@/pages/VehicleDetail'
import { DueForService } from '@/pages/DueForService'
// Lazy — pulls in the QR camera/generator libs only when the optional
// batch-scan feature is actually used.
const ScanInstall = lazy(() => import('@/pages/ScanInstall').then(m => ({ default: m.ScanInstall })))
import { VehicleReport } from '@/pages/VehicleReport'
import { Customers } from '@/pages/Customers'
import { Suppliers } from '@/pages/Suppliers'
import { SupplierDetail } from '@/pages/SupplierDetail'
import { Stocktakes } from '@/pages/Stocktakes'
import { StocktakeDetail } from '@/pages/StocktakeDetail'
import { PurchaseOrders } from '@/pages/PurchaseOrders'
import { PurchaseOrderDetail } from '@/pages/PurchaseOrderDetail'
import { Employees } from '@/pages/Employees'
import { ServiceJobDetail } from '@/pages/ServiceJobDetail'
import { ServiceJobs } from '@/pages/ServiceJobs'
import { InvoiceDetail } from '@/pages/InvoiceDetail'
import { Invoices } from '@/pages/Invoices'
import { Settings } from '@/pages/Settings'
import { Users } from '@/pages/Users'
import { AnalyticsLayout } from '@/pages/analytics/AnalyticsLayout'
import { SalesAnalytics } from '@/pages/analytics/Sales'
import { InventoryAnalytics } from '@/pages/analytics/Inventory'
import { CustomersAnalytics } from '@/pages/analytics/Customers'
import { ReceivablesAnalytics } from '@/pages/analytics/Receivables'
import { PnLAnalytics } from '@/pages/analytics/PnL'
import { AuditLog } from '@/pages/analytics/AuditLog'
import { TechniciansAnalytics } from '@/pages/analytics/Technicians'
import type { ReactNode } from 'react'

const queryClient = new QueryClient({
  defaultOptions: {
    mutations: {
      onError: (error: any) => {
        const message = error?.response?.data?.error?.message || error?.message || 'Something went wrong'
        toast.error(message)
      },
    },
  },
})

function ProtectedRoute({ children }: { children: ReactNode }) {
  const { user, loading } = useAuth()
  if (loading) return <div className="flex items-center justify-center min-h-screen text-sm text-muted-foreground">Loading...</div>
  if (!user) return <Navigate to="/login" replace />
  return <>{children}</>
}

export default function App() {
  return (
    <QueryClientProvider client={queryClient}>
      <AuthProvider>
        <BrowserRouter>
          <Routes>
            <Route path="/login" element={<Login />} />
            <Route path="/report/:token" element={<VehicleReport />} />
            <Route path="/" element={<ProtectedRoute><MainLayout /></ProtectedRoute>}>
              <Route index element={<Dashboard />} />
              <Route path="sale" element={<Sale />} />
              <Route path="inventory" element={<Inventory />} />
              <Route path="customers" element={<Customers />} />
              <Route path="customers/:id" element={<CustomerDetail />} />
              <Route path="vehicles/:id" element={<VehicleDetail />} />
              <Route path="due-for-service" element={<DueForService />} />
              <Route path="scan" element={<Suspense fallback={<p className="p-6 text-sm text-muted-foreground">Loading scanner…</p>}><ScanInstall /></Suspense>} />
              <Route path="suppliers" element={<Suppliers />} />
              <Route path="suppliers/:id" element={<SupplierDetail />} />
              <Route path="stocktakes" element={<Stocktakes />} />
              <Route path="stocktakes/:id" element={<StocktakeDetail />} />
              <Route path="purchase-orders" element={<PurchaseOrders />} />
              <Route path="purchase-orders/:id" element={<PurchaseOrderDetail />} />
              <Route path="service-jobs" element={<ServiceJobs />} />
              <Route path="service-jobs/:id" element={<ServiceJobDetail />} />
              <Route path="invoices" element={<Invoices />} />
              <Route path="invoices/:id" element={<InvoiceDetail />} />
              <Route path="reports" element={<Navigate to="/analytics/pnl" replace />} />
              <Route path="analytics" element={<AnalyticsLayout />}>
                <Route index element={<Navigate to="/analytics/sales" replace />} />
                <Route path="sales" element={<SalesAnalytics />} />
                <Route path="inventory" element={<InventoryAnalytics />} />
                <Route path="customers" element={<CustomersAnalytics />} />
                <Route path="receivables" element={<ReceivablesAnalytics />} />
                <Route path="pnl" element={<PnLAnalytics />} />
                <Route path="technicians" element={<TechniciansAnalytics />} />
                <Route path="audit" element={<AuditLog />} />
              </Route>
              <Route path="settings" element={<Settings />} />
              <Route path="users" element={<Users />} />
              <Route path="employees" element={<Employees />} />
            </Route>
          </Routes>
        </BrowserRouter>
        <Toaster position="top-right" richColors closeButton />
      </AuthProvider>
    </QueryClientProvider>
  )
}
