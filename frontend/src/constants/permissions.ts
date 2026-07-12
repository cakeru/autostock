export const PERMISSION_GROUPS = [
  {
    group: 'Inventory',
    permissions: [
      { key: 'inventory:view', label: 'View products' },
      { key: 'inventory:create', label: 'Create products' },
      { key: 'inventory:update', label: 'Update products' },
      { key: 'inventory:delete', label: 'Delete products' },
    ],
  },
  {
    group: 'Customers',
    permissions: [
      { key: 'customer:view', label: 'View customers' },
      { key: 'customer:create', label: 'Create customers' },
      { key: 'customer:update', label: 'Update customers' },
      { key: 'customer:delete', label: 'Delete customers' },
    ],
  },
  {
    group: 'Services',
    permissions: [
      { key: 'service:view', label: 'View service jobs' },
      { key: 'service:create', label: 'Create service jobs' },
      { key: 'service:update', label: 'Update service jobs' },
      { key: 'service:delete', label: 'Delete service jobs' },
    ],
  },
  {
    group: 'Invoices',
    permissions: [
      { key: 'invoice:view', label: 'View invoices' },
      { key: 'invoice:create', label: 'Create invoices' },
      { key: 'invoice:void', label: 'Void invoices' },
    ],
  },
  {
    group: 'Reports',
    permissions: [
      { key: 'report:view', label: 'View reports' },
    ],
  },
  {
    group: 'Workshop',
    permissions: [
      { key: 'install:scan', label: 'Scan batch installs (mechanic)' },
    ],
  },
]

export const DEFAULT_STAFF_PERMISSIONS = [
  'inventory:view',
  'customer:view',
  'service:view',
  'invoice:view',
  'invoice:create',
]

export const ALL_PERMISSIONS = PERMISSION_GROUPS.flatMap(g => g.permissions.map(p => p.key))
