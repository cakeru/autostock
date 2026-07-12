# Frontend Guide

## Overview

The AutoStock frontend is a React + TypeScript application built with Vite. It follows modern React patterns with functional components, hooks, and TanStack Query for data fetching.

## Tech Stack

- **Framework**: React 18+ with TypeScript
- **Build Tool**: Vite
- **State Management**: TanStack Query (React Query)
- **Form Handling**: React Hook Form + Zod
- **Styling**: Tailwind CSS
- **Component Library**: shadcn/ui (copy-paste components, not a dependency)
- **HTTP Client**: Axios
- **Routing**: React Router v6
- **Icons**: Lucide React
- **Date Handling**: date-fns
- **Charts**: Recharts (for dashboard)
- **Tables**: TanStack Table (for advanced data tables)

## Design Philosophy

### Typography
**Primary font**: Plus Jakarta Sans (Google Fonts)
- Modern, slightly geometric — balances technical precision with approachability
- Use for all UI text, headings, and data labels
- Fallback: `system-ui, -apple-system, sans-serif`
- Load via `<link>` in `index.html` or `@import` in CSS

**Monospace**: JetBrains Mono
- Use sparingly: SKU display, invoice numbers, tire size specs
- Never for body text or UI labels

### Color Palette
**Primary**: Teal-blue (`hsl(192, 91%, 36%)`)
- Distinctive from generic AI blue — feels technical and automotive
- Used for primary buttons, active nav items, links

**Accent**: Warm amber (`hsl(38, 92%, 50%)`)
- For warnings, badges, highlight states
- Complements the primary teal-blue

**Neutrals**: Warm gray scale
- Background: `hsl(210, 20%, 98%)` — very light, slightly warm
- Surface: `hsl(0, 0%, 100%)` — pure white for cards/sheets
- Text: `hsl(220, 15%, 15%)` — soft black, easier on eyes than pure black
- Muted text: `hsl(220, 10%, 55%)` — secondary information

### Border Radius
**Philosophy**: Intentional, minimal — avoids pill-shaped everything
- Buttons: `rounded` (4px)
- Inputs: `rounded` (4px) — consistent with buttons
- Cards: `rounded-md` (6px) — subtle definition
- Modals/dialogs: `rounded-md` (6px)
- Never use `rounded-lg`, `rounded-xl`, `rounded-2xl` on interactive elements

### Spacing Density
**Desktop (dense)**:
- Card padding: `p-3` (12px)
- Form field spacing: `space-y-2` (8px)
- Grid gaps: `gap-2` (8px)
- Page padding: `p-4` (16px)
- Table cells: `px-3 py-2` (12px x 8px)

**Mobile (balanced)**:
- Card padding: `p-3` or `p-4` (12-16px)
- Form field spacing: `space-y-3` (12px)
- Grid gaps: `gap-3` (12px)
- Page padding: `p-4` (16px)
- Touch targets: minimum 44x44px

### Navigation Pattern
- **Desktop (≥768px)**: Collapsible sidebar (60px collapsed / 240px expanded)
- **Mobile (<768px)**: Fixed bottom navigation bar (56px height, 5 items max)
- See Navigation Implementation section below

### Anti-Patterns (Avoid These)
- Inter or Geist fonts (too generic — every AI site uses these)
- Pure `blue-600` / `indigo-600` as primary color
- `rounded-lg`, `rounded-xl`, `rounded-2xl` on cards and buttons
- Gradient backgrounds or hero sections with headline + subheadline + CTA
- Excessive padding (`p-6` or larger) without functional reason
- Icon + label sidebar that looks like the default shadcn pattern
- Heavy shadows (`shadow-lg`, `shadow-xl`) — prefer `shadow-sm` or `border` instead
- Hover states that just change opacity — prefer background color shifts

## Component Strategy

### shadcn/ui Components (Use these)
Pre-built, accessible components from shadcn/ui. These are copied into `src/components/ui/` and fully customizable.

**Core UI:**
- Button, Input, Select, Textarea, Label
- Checkbox, RadioGroup, Switch, Slider
- Dialog (Modal), Sheet (Drawer), Popover, DropdownMenu
- Table, Pagination
- Card, Badge, Avatar
- Tabs, Accordion
- Calendar, DatePicker
- Toast (Sonner), Alert
- Command (Search/Combobox)

**When to use:**
- Standard form inputs and controls
- Modal dialogs and drawers
- Data tables with sorting/filtering
- Navigation menus and dropdowns
- Date pickers

### Custom Components (Build these)
Domain-specific components not covered by shadcn/ui.

**Dashboard:**
- RevenueChart - Line/bar chart for daily revenue
- JobsSummary - Job status breakdown
- StockAlerts - Low stock widget

**Inventory:**
- TireSpecForm - Tire size/brand/model inputs
- StockBadge - Visual stock level indicator
- ProductQuickAdd - Quick add form for common tires

**Invoice:**
- InvoicePreview - Live invoice preview
- InvoicePDF - PDF generation component
- CurrencyDisplay - Dual currency (USD/KHR) display

**Mobile:**
- MobileNav - Collapsible navigation
- SwipeableCard - Touch-friendly cards
- QuickActions - Floating action buttons

## Mobile-First Design

### Responsive Breakpoints
```css
/* Tailwind breakpoints (mobile-first) */
sm: 640px   /* Large phones */
md: 768px   /* Tablets */
lg: 1024px  /* Small laptops */
xl: 1280px  /* Desktops */
2xl: 1536px /* Large screens */
```

### Mobile Patterns

**Navigation:**
- Sidebar collapses to hamburger menu on mobile
- Bottom navigation for quick access (optional)
- Swipe gestures for navigation

**Tables:**
- Desktop: Full data table with columns
- Mobile: Card layout or horizontal scroll with sticky first column
- Use TanStack Table's responsive features

**Forms:**
- Full-width inputs on mobile
- Stack form fields vertically
- Use bottom sheets for complex forms
- Large touch targets (min 44x44px)

**Dashboard:**
- Stack widgets vertically on mobile
- Swipeable carousel for metrics
- Collapsible sections

### Touch-Friendly Guidelines
- Minimum touch target: 44x44px (11x11 Tailwind units)
- Spacing between interactive elements: 8px minimum
- Use `active:` states for tap feedback
- Avoid hover-only interactions on mobile

## shadcn/ui Setup

### Installation

```bash
cd frontend

# Initialize shadcn/ui
npx shadcn-ui@latest init

# Choose:
# - Style: New York (tighter default spacing than Default style)
# - Base color: Custom (we override this in tailwind.config.js)
# - CSS variables: Yes
```

### Custom Tailwind Config

Override the default theme in `tailwind.config.js`:

```javascript
module.exports = {
  theme: {
    extend: {
      fontFamily: {
        sans: ['Plus Jakarta Sans', 'system-ui', '-apple-system', 'sans-serif'],
        mono: ['JetBrains Mono', 'monospace'],
      },
      colors: {
        primary: {
          DEFAULT: 'hsl(192, 91%, 36%)',   // Teal-blue
          foreground: 'hsl(0, 0%, 100%)',
        },
        accent: {
          DEFAULT: 'hsl(38, 92%, 50%)',    // Warm amber
          foreground: 'hsl(0, 0%, 100%)',
        },
      },
      borderRadius: {
        DEFAULT: '4px',
        md: '6px',
      },
    },
  },
}
```

### Adding Components

```bash
# Add individual components
npx shadcn-ui@latest add button
npx shadcn-ui@latest add input
npx shadcn-ui@latest add select
npx shadcn-ui@latest add dialog
npx shadcn-ui@latest add table
npx shadcn-ui@latest add card
npx shadcn-ui@latest add badge
npx shadcn-ui@latest add toast

# Or add multiple at once
npx shadcn-ui@latest add button input select dialog table card badge toast
```

### Component Location
Components are copied to `src/components/ui/`. You own the code and can customize freely.

```
src/components/ui/
├── button.tsx
├── input.tsx
├── select.tsx
├── dialog.tsx
├── table.tsx
├── card.tsx
├── badge.tsx
└── ...
```

### Customization

**Background & surface colors (globals.css):**
```css
:root {
  --background: 210 20% 98%;
  --foreground: 220 15% 15%;
  --card: 0 0% 100%;
  --card-foreground: 220 15% 15%;
  --primary: 192 91% 36%;
  --primary-foreground: 0 0% 100%;
  --accent: 38 92% 50%;
  --accent-foreground: 0 0% 100%;
  --muted: 210 15% 95%;
  --muted-foreground: 220 10% 55%;
  --destructive: 0 84% 60%;
  --destructive-foreground: 0 0% 100%;
  --border: 220 13% 88%;
  --input: 220 13% 88%;
  --ring: 192 91% 36%;
  --radius: 0.25rem;
}
```

**Component Overrides:**
```tsx
// src/components/ui/button.tsx
// Edit the component directly to change defaults
const buttonVariants = cva(
  'inline-flex items-center justify-center rounded text-sm font-medium transition-colors',
  {
    variants: {
      variant: {
        default: 'bg-primary text-primary-foreground hover:bg-primary/90',
        destructive: 'bg-destructive text-destructive-foreground hover:bg-destructive/90',
        outline: 'border border-input bg-background hover:bg-accent/10',
        secondary: 'bg-secondary text-secondary-foreground hover:bg-secondary/80',
        ghost: 'hover:bg-accent/10 hover:text-accent-foreground',
      },
      size: {
        default: 'h-9 px-3 py-2',
        sm: 'h-8 px-2 text-xs',
        lg: 'h-10 px-4',
        icon: 'h-9 w-9',
      },
    },
  }
)
```

## Usage Examples

### Form with shadcn/ui + React Hook Form

```tsx
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { z } from 'zod'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'

const productSchema = z.object({
  name: z.string().min(1, 'Name is required'),
  sku: z.string().min(1, 'SKU is required'),
  type: z.enum(['tire', 'part', 'labor']),
  sell_price: z.number().min(0),
})

type ProductFormData = z.infer<typeof productSchema>

export const ProductForm = () => {
  const { register, handleSubmit, setValue, formState: { errors } } = useForm<ProductFormData>({
    resolver: zodResolver(productSchema),
  })

  const onSubmit = (data: ProductFormData) => {
    // Submit to API
  }

  return (
    <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
      <div className="space-y-2">
        <Label htmlFor="name">Product Name</Label>
        <Input id="name" {...register('name')} />
        {errors.name && <p className="text-sm text-destructive">{errors.name.message}</p>}
      </div>

      <div className="space-y-2">
        <Label htmlFor="sku">SKU</Label>
        <Input id="sku" {...register('sku')} />
        {errors.sku && <p className="text-sm text-destructive">{errors.sku.message}</p>}
      </div>

      <div className="space-y-2">
        <Label>Type</Label>
        <Select onValueChange={(value) => setValue('type', value as any)}>
          <SelectTrigger>
            <SelectValue placeholder="Select type" />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="tire">Tire</SelectItem>
            <SelectItem value="part">Part</SelectItem>
            <SelectItem value="labor">Labor</SelectItem>
          </SelectContent>
        </Select>
      </div>

      <div className="space-y-2">
        <Label htmlFor="sell_price">Sell Price (USD)</Label>
        <Input id="sell_price" type="number" step="0.01" {...register('sell_price', { valueAsNumber: true })} />
        {errors.sell_price && <p className="text-sm text-destructive">{errors.sell_price.message}</p>}
      </div>

      <Button type="submit" className="w-full md:w-auto">Save Product</Button>
    </form>
  )
}
```

### Navigation Implementation

**Desktop Sidebar — Collapsible:**

```tsx
// components/layout/Sidebar.tsx
import { useState } from 'react'
import { Link, useLocation } from 'react-router-dom'
import { LayoutDashboard, Package, Users, Wrench, FileText, Settings, ChevronLeft } from 'lucide-react'
import { cn } from '@/lib/utils'
import { Button } from '@/components/ui/button'

const navItems = [
  { icon: LayoutDashboard, label: 'Dashboard', path: '/' },
  { icon: Package, label: 'Inventory', path: '/inventory' },
  { icon: Users, label: 'Customers', path: '/customers' },
  { icon: Wrench, label: 'Service Jobs', path: '/service-jobs' },
  { icon: FileText, label: 'Invoices', path: '/invoices' },
  { icon: Settings, label: 'Settings', path: '/settings' },
]

export const Sidebar = () => {
  const [collapsed, setCollapsed] = useState(false)
  const location = useLocation()

  return (
    <aside className={cn(
      'fixed left-0 top-0 h-full border-r bg-background transition-all duration-200 z-30',
      collapsed ? 'w-[60px]' : 'w-[240px]'
    )}>
      <div className="flex h-14 items-center justify-between border-b px-3">
        {!collapsed && <span className="font-semibold text-sm">AutoStock</span>}
        <Button
          variant="ghost"
          size="icon"
          onClick={() => setCollapsed(!collapsed)}
        >
          <ChevronLeft className={cn('h-4 w-4 transition-transform', collapsed && 'rotate-180')} />
        </Button>
      </div>

      <nav className="flex flex-col gap-0.5 p-2">
        {navItems.map((item) => {
          const Icon = item.icon
          const isActive = location.pathname === item.path

          return (
            <Link
              key={item.path}
              to={item.path}
              className={cn(
                'flex items-center gap-3 rounded px-3 py-2 text-sm transition-colors',
                isActive
                  ? 'bg-primary/10 text-primary font-medium'
                  : 'text-muted-foreground hover:bg-accent/10 hover:text-foreground'
              )}
            >
              <Icon className="h-4 w-4 flex-shrink-0" />
              {!collapsed && <span>{item.label}</span>}
            </Link>
          )
        })}
      </nav>
    </aside>
  )
}
```

**Mobile Bottom Navigation:**

```tsx
// components/layout/MobileNav.tsx
import { Link, useLocation } from 'react-router-dom'
import { Home, Package, Wrench, FileText, MoreHorizontal } from 'lucide-react'
import { cn } from '@/lib/utils'

const navItems = [
  { icon: Home, label: 'Home', path: '/' },
  { icon: Package, label: 'Inventory', path: '/inventory' },
  { icon: Wrench, label: 'Jobs', path: '/service-jobs' },
  { icon: FileText, label: 'Invoices', path: '/invoices' },
  { icon: MoreHorizontal, label: 'More', path: '/more' },
]

export const MobileNav = () => {
  const location = useLocation()

  return (
    <nav className="fixed bottom-0 left-0 right-0 z-30 border-t bg-background md:hidden">
      <div className="flex h-14 items-center justify-around px-2">
        {navItems.map((item) => {
          const Icon = item.icon
          const isActive = location.pathname === item.path

          return (
            <Link
              key={item.path}
              to={item.path}
              className={cn(
                'flex flex-col items-center gap-0.5 px-3 py-1 text-xs min-w-0',
                isActive ? 'text-primary' : 'text-muted-foreground'
              )}
            >
              <Icon className="h-5 w-5" />
              <span className="truncate">{item.label}</span>
            </Link>
          )
        })}
      </div>
    </nav>
  )
}
```

**Responsive Layout Wrapper:**

```tsx
// components/layout/MainLayout.tsx
import { Sidebar } from './Sidebar'
import { MobileNav } from './MobileNav'

export const MainLayout = ({ children }: { children: React.ReactNode }) => {
  return (
    <div className="min-h-screen bg-background">
      {/* Desktop sidebar */}
      <div className="hidden md:block">
        <Sidebar />
      </div>

      {/* Main content area */}
      <main className="md:ml-[240px] p-4 pb-20 md:pb-4">
        {children}
      </main>

      {/* Mobile bottom nav */}
      <MobileNav />
    </div>
  )
}
```

### Data Table with TanStack Table

```tsx
import { useState } from 'react'
import {
  useReactTable,
  getCoreRowModel,
  getSortedRowModel,
  getPaginationRowModel,
  flexRender,
  ColumnDef,
} from '@tanstack/react-table'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'

interface Product {
  id: number
  name: string
  sku: string
  stock_quantity: number
  sell_price: number
}

const columns: ColumnDef<Product>[] = [
  {
    accessorKey: 'name',
    header: 'Name',
    cell: ({ row }) => <div className="font-medium">{row.getValue('name')}</div>,
  },
  {
    accessorKey: 'sku',
    header: 'SKU',
  },
  {
    accessorKey: 'stock_quantity',
    header: 'Stock',
    cell: ({ row }) => {
      const stock = row.getValue('stock_quantity') as number
      return (
        <Badge variant={stock < 5 ? 'destructive' : 'secondary'}>
          {stock}
        </Badge>
      )
    },
  },
  {
    accessorKey: 'sell_price',
    header: 'Price',
    cell: ({ row }) => {
      const price = row.getValue('sell_price') as number
      return <div>${price.toFixed(2)}</div>
    },
  },
]

export const ProductTable = ({ products }: { products: Product[] }) => {
  const [sorting, setSorting] = useState([])

  const table = useReactTable({
    data: products,
    columns,
    state: { sorting },
    onSortingChange: setSorting,
    getCoreRowModel: getCoreRowModel(),
    getSortedRowModel: getSortedRowModel(),
    getPaginationRowModel: getPaginationRowModel(),
  })

  return (
    <div className="rounded-md border">
      <Table>
        <TableHeader>
          {table.getHeaderGroups().map((headerGroup) => (
            <TableRow key={headerGroup.id}>
              {headerGroup.headers.map((header) => (
                <TableHead key={header.id}>
                  {flexRender(header.column.columnDef.header, header.getContext())}
                </TableHead>
              ))}
            </TableRow>
          ))}
        </TableHeader>
        <TableBody>
          {table.getRowModel().rows.map((row) => (
            <TableRow key={row.id}>
              {row.getVisibleCells().map((cell) => (
                <TableCell key={cell.id}>
                  {flexRender(cell.column.columnDef.cell, cell.getContext())}
                </TableCell>
              ))}
            </TableRow>
          ))}
        </TableBody>
      </Table>
    </div>
  )
}
```

## Project Structure

```
frontend/
├── public/                    # Static assets
│   ├── favicon.ico
│   └── logo.svg
├── src/
│   ├── components/           # Reusable components
│   │   ├── ui/              # Base UI components
│   │   │   ├── Button.tsx
│   │   │   ├── Input.tsx
│   │   │   ├── Select.tsx
│   │   │   ├── Modal.tsx
│   │   │   ├── Table.tsx
│   │   │   ├── Card.tsx
│   │   │   └── ...
│   │   ├── layout/          # Layout components
│   │   │   ├── Header.tsx
│   │   │   ├── Sidebar.tsx
│   │   │   ├── Footer.tsx
│   │   │   └── MainLayout.tsx
│   │   ├── dashboard/       # Dashboard widgets
│   │   │   ├── RevenueChart.tsx
│   │   │   ├── JobsSummary.tsx
│   │   │   └── StockAlerts.tsx
│   │   ├── inventory/       # Inventory components
│   │   │   ├── ProductList.tsx
│   │   │   ├── ProductForm.tsx
│   │   │   ├── TireForm.tsx
│   │   │   └── StockBadge.tsx
│   │   ├── customer/        # Customer components
│   │   │   ├── CustomerList.tsx
│   │   │   ├── CustomerForm.tsx
│   │   │   ├── VehicleList.tsx
│   │   │   └── ServiceHistory.tsx
│   │   ├── service/         # Service job components
│   │   │   ├── JobList.tsx
│   │   │   ├── JobForm.tsx
│   │   │   ├── JobStatus.tsx
│   │   │   └── JobItems.tsx
│   │   ├── invoice/         # Invoice components
│   │   │   ├── InvoiceList.tsx
│   │   │   ├── InvoiceForm.tsx
│   │   │   ├── InvoicePreview.tsx
│   │   │   └── InvoicePDF.tsx
│   │   └── settings/        # Settings components
│   │       ├── SettingsForm.tsx
│   │       ├── ExchangeRateForm.tsx
│   │       └── TelegramSettings.tsx
│   ├── hooks/               # Custom hooks
│   │   ├── useAuth.ts
│   │   ├── useProducts.ts
│   │   ├── useCustomers.ts
│   │   ├── useServiceJobs.ts
│   │   ├── useInvoices.ts
│   │   ├── useSettings.ts
│   │   └── usePermissions.ts
│   ├── pages/               # Route pages
│   │   ├── Login.tsx
│   │   ├── Dashboard.tsx
│   │   ├── Inventory.tsx
│   │   ├── Customers.tsx
│   │   ├── CustomerDetail.tsx
│   │   ├── ServiceJobs.tsx
│   │   ├── ServiceJobDetail.tsx
│   │   ├── Invoices.tsx
│   │   ├── InvoiceDetail.tsx
│   │   ├── Settings.tsx
│   │   └── Users.tsx
│   ├── services/            # API service layer
│   │   ├── api.ts           # Axios instance
│   │   ├── auth.ts
│   │   ├── products.ts
│   │   ├── customers.ts
│   │   ├── serviceJobs.ts
│   │   ├── invoices.ts
│   │   ├── settings.ts
│   │   └── dashboard.ts
│   ├── types/               # TypeScript types
│   │   ├── api.ts           # API response types
│   │   ├── product.ts
│   │   ├── customer.ts
│   │   ├── serviceJob.ts
│   │   ├── invoice.ts
│   │   ├── user.ts
│   │   └── settings.ts
│   ├── utils/               # Utility functions
│   │   ├── auth.ts          # Auth helpers
│   │   ├── currency.ts      # Currency formatting
│   │   ├── date.ts          # Date formatting
│   │   ├── validation.ts    # Zod schemas
│   │   └── permissions.ts   # Permission helpers
│   ├── contexts/            # React contexts
│   │   ├── AuthContext.tsx
│   │   └── SettingsContext.tsx
│   ├── constants/           # Constants
│   │   ├── routes.ts
│   │   ├── permissions.ts
│   │   └── config.ts
│   ├── App.tsx              # Root component
│   ├── main.tsx             # Entry point
│   └── index.css            # Global styles
├── .env.example             # Environment variables
├── index.html               # HTML template
├── package.json
├── tsconfig.json
├── vite.config.ts
├── tailwind.config.js
└── postcss.config.js
```

## Key Patterns

### 1. API Service Layer

All API calls are centralized in the `services/` directory.

```typescript
// services/api.ts
import axios from 'axios';

const api = axios.create({
  baseURL: import.meta.env.VITE_API_URL || 'http://localhost:8080/api/v1',
  headers: {
    'Content-Type': 'application/json',
  },
});

// Add auth token to requests
api.interceptors.request.use((config) => {
  const token = localStorage.getItem('access_token');
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
});

// Handle 401 errors
api.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response?.status === 401) {
      localStorage.removeItem('access_token');
      window.location.href = '/login';
    }
    return Promise.reject(error);
  }
);

export default api;
```

```typescript
// services/products.ts
import api from './api';
import { Product, ProductFilter, PaginatedResponse } from '../types';

export const productApi = {
  list: async (filter?: ProductFilter): Promise<PaginatedResponse<Product>> => {
    const response = await api.get('/products', { params: filter });
    return response.data;
  },

  get: async (id: number): Promise<Product> => {
    const response = await api.get(`/products/${id}`);
    return response.data.data;
  },

  create: async (data: Partial<Product>): Promise<Product> => {
    const response = await api.post('/products', data);
    return response.data.data;
  },

  update: async (id: number, data: Partial<Product>): Promise<Product> => {
    const response = await api.put(`/products/${id}`, data);
    return response.data.data;
  },

  delete: async (id: number): Promise<void> => {
    await api.delete(`/products/${id}`);
  },
};
```

### 2. Custom Hooks with TanStack Query

```typescript
// hooks/useProducts.ts
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { productApi } from '../services/products';
import { ProductFilter } from '../types';

export const useProducts = (filter?: ProductFilter) => {
  return useQuery({
    queryKey: ['products', filter],
    queryFn: () => productApi.list(filter),
  });
};

export const useProduct = (id: number) => {
  return useQuery({
    queryKey: ['product', id],
    queryFn: () => productApi.get(id),
    enabled: id > 0,
  });
};

export const useCreateProduct = () => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (data: Partial<Product>) => productApi.create(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['products'] });
    },
  });
};

export const useUpdateProduct = () => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ id, data }: { id: number; data: Partial<Product> }) =>
      productApi.update(id, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['products'] });
    },
  });
};

export const useDeleteProduct = () => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (id: number) => productApi.delete(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['products'] });
    },
  });
};
```

### 3. Form Handling with React Hook Form + Zod

```typescript
// components/inventory/ProductForm.tsx
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';

const productSchema = z.object({
  type: z.enum(['tire', 'part', 'labor', 'consumable']),
  sku: z.string().min(1, 'SKU is required'),
  name: z.string().min(1, 'Name is required'),
  description: z.string().optional(),
  buy_price: z.number().min(0, 'Buy price must be positive'),
  sell_price: z.number().min(0, 'Sell price must be positive'),
  stock_quantity: z.number().int().min(0, 'Stock must be positive'),
  min_stock_alert: z.number().int().min(0),
  // Tire-specific fields
  tire_size: z.string().optional(),
  tire_brand: z.string().optional(),
  tire_model: z.string().optional(),
  // ... more fields
});

type ProductFormData = z.infer<typeof productSchema>;

interface ProductFormProps {
  initialData?: Product;
  onSubmit: (data: ProductFormData) => void;
  isLoading?: boolean;
}

export const ProductForm: React.FC<ProductFormProps> = ({
  initialData,
  onSubmit,
  isLoading,
}) => {
  const {
    register,
    handleSubmit,
    setValue,
    formState: { errors },
    watch,
  } = useForm<ProductFormData>({
    resolver: zodResolver(productSchema),
    defaultValues: initialData || {
      type: 'tire',
      stock_quantity: 0,
      min_stock_alert: 5,
    },
  });

  const productType = watch('type');
  const isTire = productType === 'tire';

  return (
    <form onSubmit={handleSubmit(onSubmit)} className="space-y-2">
      <div className="space-y-1">
        <Label htmlFor="type">Type</Label>
        <Select onValueChange={(value) => setValue('type', value as any)} defaultValue={productType}>
          <SelectTrigger>
            <SelectValue placeholder="Select type" />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="tire">Tire</SelectItem>
            <SelectItem value="part">Part</SelectItem>
            <SelectItem value="labor">Labor</SelectItem>
            <SelectItem value="consumable">Consumable</SelectItem>
          </SelectContent>
        </Select>
        {errors.type && <p className="text-xs text-destructive">{errors.type.message}</p>}
      </div>

      <div className="space-y-1">
        <Label htmlFor="sku">SKU</Label>
        <Input id="sku" {...register('sku')} />
        {errors.sku && <p className="text-xs text-destructive">{errors.sku.message}</p>}
      </div>

      <div className="space-y-1">
        <Label htmlFor="name">Name</Label>
        <Input id="name" {...register('name')} />
        {errors.name && <p className="text-xs text-destructive">{errors.name.message}</p>}
      </div>

      {/* Tire-specific fields — only show when type is tire */}
      {isTire && (
        <div className="border rounded-md p-3 space-y-2">
          <Label className="text-xs text-muted-foreground uppercase tracking-wide">Tire Specs</Label>
          <div className="grid grid-cols-2 gap-2">
            <div className="space-y-1">
              <Label htmlFor="tire_size" className="text-xs">Size</Label>
              <Input id="tire_size" {...register('tire_size')} placeholder="205/55R16" />
            </div>
            <div className="space-y-1">
              <Label htmlFor="tire_brand" className="text-xs">Brand</Label>
              <Input id="tire_brand" {...register('tire_brand')} />
            </div>
          </div>
        </div>
      )}

      <div className="flex gap-2 pt-2">
        <Button type="submit" disabled={isLoading}>
          {isLoading ? 'Saving...' : 'Save Product'}
        </Button>
      </div>
    </form>
  );
};
```

### 4. Permission-Based UI

```typescript
// contexts/AuthContext.tsx
import { createContext, useContext, useState, useEffect } from 'react';

interface AuthContextType {
  user: User | null;
  login: (username: string, password: string) => Promise<void>;
  logout: () => void;
  hasPermission: (permission: string) => boolean;
}

const AuthContext = createContext<AuthContextType | null>(null);

export const AuthProvider: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const [user, setUser] = useState<User | null>(null);

  useEffect(() => {
    const token = localStorage.getItem('access_token');
    if (token) {
      // Fetch user data
      authApi.getMe().then(setUser);
    }
  }, []);

  const hasPermission = (permission: string) => {
    if (!user) return false;
    if (user.role === 'admin') return true;
    return user.permissions.includes(permission);
  };

  return (
    <AuthContext.Provider value={{ user, login, logout, hasPermission }}>
      {children}
    </AuthContext.Provider>
  );
};

export const useAuth = () => {
  const context = useContext(AuthContext);
  if (!context) {
    throw new Error('useAuth must be used within AuthProvider');
  }
  return context;
};
```

```typescript
// components/ui/PermissionGuard.tsx
import { useAuth } from '../../contexts/AuthContext';

interface PermissionGuardProps {
  permission: string;
  children: React.ReactNode;
  fallback?: React.ReactNode;
}

export const PermissionGuard: React.FC<PermissionGuardProps> = ({
  permission,
  children,
  fallback = null,
}) => {
  const { hasPermission } = useAuth();

  if (!hasPermission(permission)) {
    return <>{fallback}</>;
  }

  return <>{children}</>;
};

// Usage
<PermissionGuard permission="inventory:create">
  <Button>Create Product</Button>
</PermissionGuard>
```

### 5. Currency Formatting

```typescript
// utils/currency.ts
export const formatUSD = (amount: number): string => {
  return new Intl.NumberFormat('en-US', {
    style: 'currency',
    currency: 'USD',
  }).format(amount);
};

export const formatKHR = (amount: number): string => {
  return new Intl.NumberFormat('km-KH', {
    style: 'currency',
    currency: 'KHR',
    maximumFractionDigits: 0,
  }).format(amount);
};

export const formatCurrency = (amount: number, currency: 'USD' | 'KHR'): string => {
  return currency === 'USD' ? formatUSD(amount) : formatKHR(amount);
};

export const convertToKHR = (usdAmount: number, exchangeRate: number): number => {
  return usdAmount * exchangeRate;
};
```

### 6. Date Formatting

```typescript
// utils/date.ts
import { format, parseISO, formatDistanceToNow } from 'date-fns';

export const formatDate = (date: string | Date): string => {
  const d = typeof date === 'string' ? parseISO(date) : date;
  return format(d, 'MMM d, yyyy');
};

export const formatDateTime = (date: string | Date): string => {
  const d = typeof date === 'string' ? parseISO(date) : date;
  return format(d, 'MMM d, yyyy HH:mm');
};

export const formatRelativeTime = (date: string | Date): string => {
  const d = typeof date === 'string' ? parseISO(date) : date;
  return formatDistanceToNow(d, { addSuffix: true });
};
```

## Routing

```typescript
// App.tsx
import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom';
import { AuthProvider, useAuth } from './contexts/AuthContext';
import { MainLayout } from './components/layout/MainLayout';
import { Login } from './pages/Login';
import { Dashboard } from './pages/Dashboard';
import { Inventory } from './pages/Inventory';
import { Customers } from './pages/Customers';
import { CustomerDetail } from './pages/CustomerDetail';
import { ServiceJobs } from './pages/ServiceJobs';
import { ServiceJobDetail } from './pages/ServiceJobDetail';
import { Invoices } from './pages/Invoices';
import { InvoiceDetail } from './pages/InvoiceDetail';
import { Settings } from './pages/Settings';
import { Users } from './pages/Users';

const ProtectedRoute: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const { user } = useAuth();
  
  if (!user) {
    return <Navigate to="/login" replace />;
  }
  
  return <>{children}</>;
};

const App: React.FC = () => {
  return (
    <AuthProvider>
      <BrowserRouter>
        <Routes>
          <Route path="/login" element={<Login />} />
          
          <Route
            path="/"
            element={
              <ProtectedRoute>
                <MainLayout />
              </ProtectedRoute>
            }
          >
            <Route index element={<Dashboard />} />
            <Route path="inventory" element={<Inventory />} />
            <Route path="customers" element={<Customers />} />
            <Route path="customers/:id" element={<CustomerDetail />} />
            <Route path="service-jobs" element={<ServiceJobs />} />
            <Route path="service-jobs/:id" element={<ServiceJobDetail />} />
            <Route path="invoices" element={<Invoices />} />
            <Route path="invoices/:id" element={<InvoiceDetail />} />
            <Route path="settings" element={<Settings />} />
            <Route path="users" element={<Users />} />
          </Route>
        </Routes>
      </BrowserRouter>
    </AuthProvider>
  );
};

export default App;
```

## State Management

### Local State
- Use `useState` for simple component state
- Use `useReducer` for complex state logic

### Server State
- Use TanStack Query for all API data
- Cache keys follow the pattern: `['resource', id, filters]`

### Global State
- Use React Context for auth and settings
- Avoid global state for data that can be fetched from API

## Styling

### Tailwind CSS

```typescript
// Example component with proper spacing and radius
// Dense desktop padding with balanced mobile padding
const Card: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  return (
    <div className="bg-card border rounded-md p-3 md:p-3">
      {children}
    </div>
  );
};

// Responsive grid with consistent gaps
<div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-2 md:gap-3">
  {/* Cards */}
</div>

// Compact table cells
<td className="px-3 py-2 text-sm">{value}</td>

// Dense form layout
<form className="space-y-2">
  <div className="space-y-1">
    {/* form fields */}
  </div>
</form>

// Data value display (SKU, invoice numbers)
<span className="font-mono text-xs text-muted-foreground">{sku}</span>
```

### Custom Components

Build reusable UI components in `components/ui/`:

```typescript
// components/ui/Button.tsx
import { ButtonHTMLAttributes, forwardRef } from 'react';

interface ButtonProps extends ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: 'primary' | 'secondary' | 'outline' | 'ghost';
  size?: 'sm' | 'md' | 'lg';
}

export const Button = forwardRef<HTMLButtonElement, ButtonProps>(
  ({ variant = 'primary', size = 'md', className = '', children, ...props }, ref) => {
    const baseStyles = 'inline-flex items-center justify-center rounded text-sm font-medium transition-colors focus:outline-none focus:ring-2 focus:ring-ring';
    
    const variants = {
      primary: 'bg-primary text-primary-foreground hover:bg-primary/90 focus:ring-primary',
      secondary: 'bg-secondary text-secondary-foreground hover:bg-secondary/80 focus:ring-secondary',
      outline: 'border border-input bg-background hover:bg-accent/10 focus:ring-primary',
      ghost: 'hover:bg-accent/10 hover:text-accent-foreground focus:ring-primary',
    };
    
    const sizes = {
      sm: 'h-8 px-2 text-xs',
      md: 'h-9 px-3',
      lg: 'h-10 px-4',
    };
    
    return (
      <button
        ref={ref}
        className={`${baseStyles} ${variants[variant]} ${sizes[size]} ${className}`}
        {...props}
      >
        {children}
      </button>
    );
  }
);
```

## Error Handling

### API Errors

```typescript
// services/api.ts
api.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response) {
      const { status, data } = error.response;
      
      switch (status) {
        case 400:
          console.error('Bad request:', data.error);
          break;
        case 401:
          localStorage.removeItem('access_token');
          window.location.href = '/login';
          break;
        case 403:
          console.error('Forbidden:', data.error);
          break;
        case 404:
          console.error('Not found:', data.error);
          break;
        case 500:
          console.error('Server error:', data.error);
          break;
      }
    }
    return Promise.reject(error);
  }
);
```

### Error Boundaries

```typescript
// components/ErrorBoundary.tsx
import { Component, ErrorInfo, ReactNode } from 'react';

interface Props {
  children: ReactNode;
  fallback?: ReactNode;
}

interface State {
  hasError: boolean;
}

export class ErrorBoundary extends Component<Props, State> {
  state: State = { hasError: false };

  static getDerivedStateFromError(): State {
    return { hasError: true };
  }

  componentDidCatch(error: Error, errorInfo: ErrorInfo) {
    console.error('Error caught by boundary:', error, errorInfo);
  }

  render() {
    if (this.state.hasError) {
      return this.props.fallback || <div>Something went wrong</div>;
    }
    return this.props.children;
  }
}
```

## Testing

### Unit Tests (Vitest)

```typescript
// utils/currency.test.ts
import { describe, it, expect } from 'vitest';
import { formatUSD, formatKHR, convertToKHR } from './currency';

describe('currency utils', () => {
  it('formats USD correctly', () => {
    expect(formatUSD(1234.56)).toBe('$1,234.56');
  });

  it('formats KHR correctly', () => {
    expect(formatKHR(1000000)).toBe('៛1,000,000');
  });

  it('converts USD to KHR', () => {
    expect(convertToKHR(100, 4050)).toBe(405000);
  });
});
```

### Component Tests (React Testing Library)

```typescript
// components/ui/Button.test.tsx
import { render, screen, fireEvent } from '@testing-library/react';
import { describe, it, expect, vi } from 'vitest';
import { Button } from './Button';

describe('Button', () => {
  it('renders correctly', () => {
    render(<Button>Click me</Button>);
    expect(screen.getByText('Click me')).toBeInTheDocument();
  });

  it('calls onClick when clicked', () => {
    const handleClick = vi.fn();
    render(<Button onClick={handleClick}>Click me</Button>);
    
    fireEvent.click(screen.getByText('Click me'));
    expect(handleClick).toHaveBeenCalledTimes(1);
  });
});
```

## Performance Optimization

### Code Splitting

```typescript
// Lazy load routes
const Dashboard = lazy(() => import('./pages/Dashboard'));
const Inventory = lazy(() => import('./pages/Inventory'));

// Use Suspense
<Suspense fallback={<div>Loading...</div>}>
  <Dashboard />
</Suspense>
```

### Memoization

```typescript
import { memo, useMemo, useCallback } from 'react';

// Memoize expensive components
const ProductList = memo(({ products }: { products: Product[] }) => {
  return (
    <ul>
      {products.map(p => <li key={p.id}>{p.name}</li>)}
    </ul>
  );
});

// Memoize expensive calculations
const filteredProducts = useMemo(() => {
  return products.filter(p => p.stock_quantity < p.min_stock_alert);
}, [products]);

// Memoize callbacks
const handleClick = useCallback((id: number) => {
  setSelectedId(id);
}, []);
```

### Virtualization

For long lists, use `react-window` or `react-virtualized`:

```typescript
import { FixedSizeList } from 'react-window';

<FixedSizeList
  height={600}
  itemCount={items.length}
  itemSize={50}
  width="100%"
>
  {({ index, style }) => (
    <div style={style}>
      {items[index].name}
    </div>
  )}
</FixedSizeList>
```

## Accessibility

- Use semantic HTML elements
- Add ARIA labels where needed
- Ensure keyboard navigation works
- Maintain sufficient color contrast
- Test with screen readers

```typescript
// Accessible form
<label htmlFor="name-input">Name</label>
<input id="name-input" type="text" aria-describedby="name-error" />
{error && <span id="name-error" role="alert">{error}</span>}
```

## Internationalization (Future)

Use `react-i18next` for Khmer language support:

```typescript
// i18n.ts
import i18n from 'i18next';
import { useTranslation } from 'react-i18next';

i18n.init({
  resources: {
    en: { translation: { /* English translations */ } },
    km: { translation: { /* Khmer translations */ } },
  },
  lng: 'en',
  fallbackLng: 'en',
});

// Usage
const { t } = useTranslation();
<h1>{t('dashboard.title')}</h1>
```

## Build & Deployment

### Development

```bash
npm run dev
```

### Production Build

```bash
npm run build
```

### Environment Variables

```bash
# .env
VITE_API_URL=http://localhost:8080/api/v1
VITE_APP_NAME=AutoStock
```

### Docker

```dockerfile
# Dockerfile
FROM node:18-alpine AS builder
WORKDIR /app
COPY package*.json ./
RUN npm ci
COPY . .
RUN npm run build

FROM nginx:alpine
COPY --from=builder /app/dist /usr/share/nginx/html
COPY nginx.conf /etc/nginx/conf.d/default.conf
EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]
```
