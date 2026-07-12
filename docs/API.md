# API Documentation

## Overview

AutoStock provides a RESTful API for all operations. The API follows standard conventions:
- Base URL: `/api/v1`
- Authentication: JWT Bearer token
- Content-Type: `application/json`
- Response format: Consistent JSON structure

## Authentication

All API endpoints (except login) require authentication via JWT token.

### Login

**Endpoint:** `POST /api/v1/auth/login`

**Request:**
```json
{
  "username": "admin",
  "password": "admin123"
}
```

**Response (200 OK):**
```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "token_type": "Bearer",
  "expires_in": 3600,
  "user": {
    "id": 1,
    "username": "admin",
    "full_name": "Administrator",
    "role": "admin",
    "permissions": ["inventory:view", "inventory:create", "..."]
  }
}
```

**Error (401 Unauthorized):**
```json
{
  "error": {
    "code": "INVALID_CREDENTIALS",
    "message": "Invalid username or password"
  }
}
```

### Refresh Token

**Endpoint:** `POST /api/v1/auth/refresh`

**Request:**
```json
{
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

**Response (200 OK):**
```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "token_type": "Bearer",
  "expires_in": 3600
}
```

### Get Current User

**Endpoint:** `GET /api/v1/auth/me`

**Headers:**
```
Authorization: Bearer <access_token>
```

**Response (200 OK):**
```json
{
  "id": 1,
  "username": "admin",
  "email": "admin@autostock.com",
  "full_name": "Administrator",
  "role": "admin",
  "permissions": ["inventory:view", "inventory:create", "..."],
  "branch_id": 1,
  "last_login_at": "2026-07-04T10:30:00Z"
}
```

### Logout

**Endpoint:** `POST /api/v1/auth/logout`

**Headers:**
```
Authorization: Bearer <access_token>
```

**Response (200 OK):**
```json
{
  "message": "Logged out successfully"
}
```

## Response Format

### Success Response
```json
{
  "data": {
    // Response data
  },
  "meta": {
    "page": 1,
    "per_page": 20,
    "total": 100,
    "total_pages": 5
  }
}
```

### Error Response
```json
{
  "error": {
    "code": "ERROR_CODE",
    "message": "Human-readable error message",
    "details": {
      // Additional error details
    }
  }
}
```

### Common Error Codes
- `INVALID_REQUEST` - Request validation failed
- `UNAUTHORIZED` - Authentication required
- `FORBIDDEN` - Insufficient permissions
- `NOT_FOUND` - Resource not found
- `CONFLICT` - Resource already exists
- `INTERNAL_ERROR` - Server error

## Pagination

List endpoints support pagination via query parameters:

**Query Parameters:**
- `page` - Page number (default: 1)
- `per_page` - Items per page (default: 20, max: 100)
- `sort_by` - Sort field (e.g., `created_at`, `name`)
- `sort_order` - Sort direction (`asc` or `desc`, default: `desc`)

**Example:**
```
GET /api/v1/products?page=2&per_page=50&sort_by=name&sort_order=asc
```

## Filtering

List endpoints support filtering via query parameters:

**Example:**
```
GET /api/v1/products?type=tire&stock_quantity_lt=10&is_active=true
```

**Common Filter Operators:**
- `field=value` - Exact match
- `field_lt=value` - Less than
- `field_gt=value` - Greater than
- `field_lte=value` - Less than or equal
- `field_gte=value` - Greater than or equal
- `field_like=value` - Contains (case-insensitive)
- `field_in=v1,v2,v3` - In list

## Users

### List Users

**Endpoint:** `GET /api/v1/users`

**Permissions:** `user:view`

**Query Parameters:**
- `role` - Filter by role (`admin` or `staff`)
- `is_active` - Filter by status (`true` or `false`)

**Response (200 OK):**
```json
{
  "data": [
    {
      "id": 1,
      "username": "admin",
      "email": "admin@autostock.com",
      "full_name": "Administrator",
      "role": "admin",
      "permissions": ["inventory:view", "..."],
      "is_active": true,
      "last_login_at": "2026-07-04T10:30:00Z",
      "created_at": "2026-01-01T00:00:00Z"
    }
  ],
  "meta": {
    "page": 1,
    "per_page": 20,
    "total": 3,
    "total_pages": 1
  }
}
```

### Get User

**Endpoint:** `GET /api/v1/users/:id`

**Permissions:** `user:view`

**Response (200 OK):**
```json
{
  "data": {
    "id": 1,
    "username": "admin",
    "email": "admin@autostock.com",
    "full_name": "Administrator",
    "role": "admin",
    "permissions": ["inventory:view", "inventory:create", "..."],
    "is_active": true,
    "last_login_at": "2026-07-04T10:30:00Z",
    "created_at": "2026-01-01T00:00:00Z",
    "updated_at": "2026-07-04T10:30:00Z"
  }
}
```

### Create User

**Endpoint:** `POST /api/v1/users`

**Permissions:** `user:create`

**Request:**
```json
{
  "username": "staff1",
  "password": "securepassword123",
  "email": "staff1@autostock.com",
  "full_name": "Staff Member",
  "role": "staff",
  "permissions": [
    "inventory:view",
    "inventory:create",
    "inventory:update",
    "customer:view",
    "customer:create",
    "customer:update",
    "service:view",
    "service:create",
    "service:update",
    "invoice:view",
    "invoice:create"
  ]
}
```

**Response (201 Created):**
```json
{
  "data": {
    "id": 2,
    "username": "staff1",
    "email": "staff1@autostock.com",
    "full_name": "Staff Member",
    "role": "staff",
    "permissions": ["inventory:view", "..."],
    "is_active": true,
    "created_at": "2026-07-04T12:00:00Z"
  }
}
```

### Update User

**Endpoint:** `PUT /api/v1/users/:id`

**Permissions:** `user:update`

**Request:**
```json
{
  "email": "newemail@autostock.com",
  "full_name": "Updated Name",
  "permissions": ["inventory:view", "inventory:create"],
  "is_active": true
}
```

**Response (200 OK):**
```json
{
  "data": {
    "id": 2,
    "username": "staff1",
    "email": "newemail@autostock.com",
    "full_name": "Updated Name",
    "role": "staff",
    "permissions": ["inventory:view", "inventory:create"],
    "is_active": true,
    "updated_at": "2026-07-04T12:30:00Z"
  }
}
```

### Change Password

**Endpoint:** `PUT /api/v1/users/:id/password`

**Permissions:** `user:update` or own account

**Request:**
```json
{
  "current_password": "oldpassword",
  "new_password": "newpassword123"
}
```

**Response (200 OK):**
```json
{
  "message": "Password updated successfully"
}
```

### Delete User

**Endpoint:** `DELETE /api/v1/users/:id`

**Permissions:** `user:delete`

**Response (204 No Content)**

## Inventory

### List Products

**Endpoint:** `GET /api/v1/products`

**Permissions:** `inventory:view`

**Query Parameters:**
- `type` - Filter by type (`tire`, `part`, `labor`, `consumable`)
- `sku` - Filter by SKU
- `name_like` - Search by name (case-insensitive)
- `category` - Filter by category
- `tire_size` - Filter by tire size
- `tire_brand` - Filter by tire brand
- `stock_quantity_lt` - Low stock filter
- `is_active` - Filter by status

**Response (200 OK):**
```json
{
  "data": [
    {
      "id": 1,
      "type": "tire",
      "sku": "TIRE-MIC-20555R16",
      "name": "Michelin Primacy 4 205/55R16",
      "description": "Premium passenger tire",
      "category": "tires",
      "buy_price": 80.00,
      "sell_price": 120.00,
      "stock_quantity": 15,
      "min_stock_alert": 5,
      "unit": "piece",
      "tire_size": "205/55R16",
      "tire_brand": "Michelin",
      "tire_model": "Primacy 4",
      "tire_pattern": "Symmetric",
      "dot_code": "2026",
      "load_index": "91",
      "speed_rating": "V",
      "tire_type": "passenger",
      "location": "A-01-03",
      "is_active": true,
      "created_at": "2026-01-15T10:00:00Z",
      "updated_at": "2026-07-04T08:00:00Z"
    }
  ],
  "meta": {
    "page": 1,
    "per_page": 20,
    "total": 45,
    "total_pages": 3
  }
}
```

### Get Product

**Endpoint:** `GET /api/v1/products/:id`

**Permissions:** `inventory:view`

**Response (200 OK):**
```json
{
  "data": {
    "id": 1,
    "type": "tire",
    "sku": "TIRE-MIC-20555R16",
    "name": "Michelin Primacy 4 205/55R16",
    "description": "Premium passenger tire",
    "category": "tires",
    "buy_price": 80.00,
    "sell_price": 120.00,
    "stock_quantity": 15,
    "min_stock_alert": 5,
    "unit": "piece",
    "tire_size": "205/55R16",
    "tire_brand": "Michelin",
    "tire_model": "Primacy 4",
    "tire_pattern": "Symmetric",
    "dot_code": "2026",
    "load_index": "91",
    "speed_rating": "V",
    "tire_type": "passenger",
    "location": "A-01-03",
    "is_active": true,
    "created_at": "2026-01-15T10:00:00Z",
    "updated_at": "2026-07-04T08:00:00Z"
  }
}
```

### Create Product

**Endpoint:** `POST /api/v1/products`

**Permissions:** `inventory:create`

**Request (Tire):**
```json
{
  "type": "tire",
  "sku": "TIRE-MIC-20555R16",
  "name": "Michelin Primacy 4 205/55R16",
  "description": "Premium passenger tire",
  "category": "tires",
  "buy_price": 80.00,
  "sell_price": 120.00,
  "stock_quantity": 15,
  "min_stock_alert": 5,
  "unit": "piece",
  "tire_size": "205/55R16",
  "tire_brand": "Michelin",
  "tire_model": "Primacy 4",
  "tire_pattern": "Symmetric",
  "dot_code": "2026",
  "load_index": "91",
  "speed_rating": "V",
  "tire_type": "passenger",
  "location": "A-01-03"
}
```

**Request (Labor):**
```json
{
  "type": "labor",
  "sku": "LABOR-TIRE-INSTALL",
  "name": "Tire Installation",
  "description": "Mount and balance one tire",
  "category": "services",
  "buy_price": 0.00,
  "sell_price": 15.00,
  "stock_quantity": 999,
  "min_stock_alert": 0,
  "unit": "hour"
}
```

**Response (201 Created):**
```json
{
  "data": {
    "id": 1,
    "type": "tire",
    "sku": "TIRE-MIC-20555R16",
    "name": "Michelin Primacy 4 205/55R16",
    "description": "Premium passenger tire",
    "category": "tires",
    "buy_price": 80.00,
    "sell_price": 120.00,
    "stock_quantity": 15,
    "min_stock_alert": 5,
    "unit": "piece",
    "tire_size": "205/55R16",
    "tire_brand": "Michelin",
    "tire_model": "Primacy 4",
    "tire_pattern": "Symmetric",
    "dot_code": "2026",
    "load_index": "91",
    "speed_rating": "V",
    "tire_type": "passenger",
    "location": "A-01-03",
    "is_active": true,
    "created_at": "2026-07-04T12:00:00Z"
  }
}
```

### Update Product

**Endpoint:** `PUT /api/v1/products/:id`

**Permissions:** `inventory:update`

**Request:**
```json
{
  "name": "Updated Product Name",
  "sell_price": 125.00,
  "stock_quantity": 20,
  "location": "A-02-01"
}
```

**Response (200 OK):**
```json
{
  "data": {
    "id": 1,
    "type": "tire",
    "sku": "TIRE-MIC-20555R16",
    "name": "Updated Product Name",
    "description": "Premium passenger tire",
    "category": "tires",
    "buy_price": 80.00,
    "sell_price": 125.00,
    "stock_quantity": 20,
    "min_stock_alert": 5,
    "unit": "piece",
    "tire_size": "205/55R16",
    "tire_brand": "Michelin",
    "tire_model": "Primacy 4",
    "tire_pattern": "Symmetric",
    "dot_code": "2026",
    "load_index": "91",
    "speed_rating": "V",
    "tire_type": "passenger",
    "location": "A-02-01",
    "is_active": true,
    "updated_at": "2026-07-04T12:30:00Z"
  }
}
```

### Delete Product

**Endpoint:** `DELETE /api/v1/products/:id`

**Permissions:** `inventory:delete`

**Response (204 No Content)**

### Get Low Stock Products

**Endpoint:** `GET /api/v1/products/low-stock`

**Permissions:** `inventory:view`

**Response (200 OK):**
```json
{
  "data": [
    {
      "id": 5,
      "sku": "TIRE-BRI-19565R15",
      "name": "Bridgestone Turanza 195/65R15",
      "stock_quantity": 3,
      "min_stock_alert": 5,
      "sell_price": 95.00
    }
  ]
}
```

## Customers

### List Customers

**Endpoint:** `GET /api/v1/customers`

**Permissions:** `customer:view`

**Query Parameters:**
- `name_like` - Search by name
- `phone` - Filter by phone
- `email` - Filter by email
- `is_active` - Filter by status

**Response (200 OK):**
```json
{
  "data": [
    {
      "id": 1,
      "name": "John Doe",
      "phone": "+855 12 345 678",
      "email": "john.doe@example.com",
      "address": "Phnom Penh, Cambodia",
      "notes": "Regular customer",
      "customer_since": "2025-03-15",
      "is_active": true,
      "created_at": "2025-03-15T10:00:00Z",
      "updated_at": "2026-07-04T08:00:00Z",
      "vehicle_count": 2
    }
  ],
  "meta": {
    "page": 1,
    "per_page": 20,
    "total": 150,
    "total_pages": 8
  }
}
```

### Get Customer

**Endpoint:** `GET /api/v1/customers/:id`

**Permissions:** `customer:view`

**Response (200 OK):**
```json
{
  "data": {
    "id": 1,
    "name": "John Doe",
    "phone": "+855 12 345 678",
    "email": "john.doe@example.com",
    "address": "Phnom Penh, Cambodia",
    "notes": "Regular customer",
    "customer_since": "2025-03-15",
    "is_active": true,
    "created_at": "2025-03-15T10:00:00Z",
    "updated_at": "2026-07-04T08:00:00Z",
    "vehicles": [
      {
        "id": 1,
        "plate_number": "ABC-1234",
        "make": "Toyota",
        "model": "Camry",
        "year": 2020,
        "vin": "1HGCM82633A123456",
        "color": "Silver"
      }
    ]
  }
}
```

### Create Customer

**Endpoint:** `POST /api/v1/customers`

**Permissions:** `customer:create`

**Request:**
```json
{
  "name": "John Doe",
  "phone": "+855 12 345 678",
  "email": "john.doe@example.com",
  "address": "Phnom Penh, Cambodia",
  "notes": "Regular customer"
}
```

**Response (201 Created):**
```json
{
  "data": {
    "id": 1,
    "name": "John Doe",
    "phone": "+855 12 345 678",
    "email": "john.doe@example.com",
    "address": "Phnom Penh, Cambodia",
    "notes": "Regular customer",
    "customer_since": "2026-07-04",
    "is_active": true,
    "created_at": "2026-07-04T12:00:00Z"
  }
}
```

### Update Customer

**Endpoint:** `PUT /api/v1/customers/:id`

**Permissions:** `customer:update`

**Request:**
```json
{
  "phone": "+855 98 765 432",
  "email": "newemail@example.com",
  "notes": "Updated notes"
}
```

**Response (200 OK):**
```json
{
  "data": {
    "id": 1,
    "name": "John Doe",
    "phone": "+855 98 765 432",
    "email": "newemail@example.com",
    "address": "Phnom Penh, Cambodia",
    "notes": "Updated notes",
    "customer_since": "2025-03-15",
    "is_active": true,
    "updated_at": "2026-07-04T12:30:00Z"
  }
}
```

### Delete Customer

**Endpoint:** `DELETE /api/v1/customers/:id`

**Permissions:** `customer:delete`

**Response (204 No Content)**

### Get Customer Service History

**Endpoint:** `GET /api/v1/customers/:id/history`

**Permissions:** `customer:view`

**Response (200 OK):**
```json
{
  "data": [
    {
      "id": 1,
      "job_number": "JOB-2026-0001",
      "status": "completed",
      "description": "Tire replacement and alignment",
      "work_performed": "Replaced 2 front tires, performed wheel alignment",
      "total_amount": 255.00,
      "completed_at": "2026-07-01T14:30:00Z"
    },
    {
      "id": 2,
      "job_number": "JOB-2026-0015",
      "status": "completed",
      "description": "Oil change",
      "work_performed": "Changed engine oil and oil filter",
      "total_amount": 45.00,
      "completed_at": "2026-06-15T11:00:00Z"
    }
  ]
}
```

## Vehicles

### List Customer Vehicles

**Endpoint:** `GET /api/v1/customers/:customer_id/vehicles`

**Permissions:** `customer:view`

**Response (200 OK):**
```json
{
  "data": [
    {
      "id": 1,
      "customer_id": 1,
      "plate_number": "ABC-1234",
      "make": "Toyota",
      "model": "Camry",
      "year": 2020,
      "vin": "1HGCM82633A123456",
      "color": "Silver",
      "notes": "Primary vehicle",
      "created_at": "2025-03-15T10:00:00Z"
    }
  ]
}
```

### Create Vehicle

**Endpoint:** `POST /api/v1/customers/:customer_id/vehicles`

**Permissions:** `customer:create`

**Request:**
```json
{
  "plate_number": "ABC-1234",
  "make": "Toyota",
  "model": "Camry",
  "year": 2020,
  "vin": "1HGCM82633A123456",
  "color": "Silver",
  "notes": "Primary vehicle"
}
```

**Response (201 Created):**
```json
{
  "data": {
    "id": 1,
    "customer_id": 1,
    "plate_number": "ABC-1234",
    "make": "Toyota",
    "model": "Camry",
    "year": 2020,
    "vin": "1HGCM82633A123456",
    "color": "Silver",
    "notes": "Primary vehicle",
    "created_at": "2026-07-04T12:00:00Z"
  }
}
```

### Update Vehicle

**Endpoint:** `PUT /api/v1/vehicles/:id`

**Permissions:** `customer:update`

**Request:**
```json
{
  "color": "Blue",
  "notes": "Updated notes"
}
```

**Response (200 OK):**
```json
{
  "data": {
    "id": 1,
    "customer_id": 1,
    "plate_number": "ABC-1234",
    "make": "Toyota",
    "model": "Camry",
    "year": 2020,
    "vin": "1HGCM82633A123456",
    "color": "Blue",
    "notes": "Updated notes",
    "updated_at": "2026-07-04T12:30:00Z"
  }
}
```

### Delete Vehicle

**Endpoint:** `DELETE /api/v1/vehicles/:id`

**Permissions:** `customer:delete`

**Response (204 No Content)**

## Service Jobs

### List Service Jobs

**Endpoint:** `GET /api/v1/service-jobs`

**Permissions:** `service:view`

**Query Parameters:**
- `status` - Filter by status (`pending`, `in_progress`, `completed`, `cancelled`)
- `customer_id` - Filter by customer
- `vehicle_id` - Filter by vehicle
- `created_at_gte` - Filter by date range
- `created_at_lte` - Filter by date range

**Response (200 OK):**
```json
{
  "data": [
    {
      "id": 1,
      "branch_id": 1,
      "job_number": "JOB-2026-0001",
      "status": "completed",
      "priority": "normal",
      "customer": {
        "id": 1,
        "name": "John Doe",
        "phone": "+855 12 345 678"
      },
      "vehicle": {
        "id": 1,
        "plate_number": "ABC-1234",
        "make": "Toyota",
        "model": "Camry"
      },
      "description": "Tire replacement and alignment",
      "diagnosis": "Front tires worn, alignment off",
      "work_performed": "Replaced 2 front tires, performed wheel alignment",
      "estimated_hours": 2.0,
      "actual_hours": 1.5,
      "started_at": "2026-07-01T13:00:00Z",
      "completed_at": "2026-07-01T14:30:00Z",
      "invoice_id": 1,
      "created_at": "2026-07-01T13:00:00Z"
    }
  ],
  "meta": {
    "page": 1,
    "per_page": 20,
    "total": 25,
    "total_pages": 2
  }
}
```

### Get Service Job

**Endpoint:** `GET /api/v1/service-jobs/:id`

**Permissions:** `service:view`

**Response (200 OK):**
```json
{
  "data": {
    "id": 1,
    "branch_id": 1,
    "job_number": "JOB-2026-0001",
    "status": "completed",
    "priority": "normal",
    "customer": {
      "id": 1,
      "name": "John Doe",
      "phone": "+855 12 345 678"
    },
    "vehicle": {
      "id": 1,
      "plate_number": "ABC-1234",
      "make": "Toyota",
      "model": "Camry",
      "year": 2020
    },
    "description": "Tire replacement and alignment",
    "diagnosis": "Front tires worn, alignment off",
    "work_performed": "Replaced 2 front tires, performed wheel alignment",
    "estimated_hours": 2.0,
    "actual_hours": 1.5,
    "started_at": "2026-07-01T13:00:00Z",
    "completed_at": "2026-07-01T14:30:00Z",
    "invoice_id": 1,
    "notes": "Customer satisfied with work",
    "items": [
      {
        "id": 1,
        "product_id": 1,
        "description": "Michelin Primacy 4 205/55R16",
        "quantity": 2,
        "unit_price": 120.00,
        "total_price": 240.00
      },
      {
        "id": 2,
        "product_id": 10,
        "description": "Wheel Alignment",
        "quantity": 1,
        "unit_price": 15.00,
        "total_price": 15.00
      }
    ],
    "created_at": "2026-07-01T13:00:00Z",
    "updated_at": "2026-07-01T14:30:00Z"
  }
}
```

### Create Service Job

**Endpoint:** `POST /api/v1/service-jobs`

**Permissions:** `service:create`

**Request:**
```json
{
  "customer_id": 1,
  "vehicle_id": 1,
  "priority": "normal",
  "description": "Tire replacement and alignment",
  "estimated_hours": 2.0,
  "notes": "Customer requested Michelin tires"
}
```

**Response (201 Created):**
```json
{
  "data": {
    "id": 1,
    "branch_id": 1,
    "job_number": "JOB-2026-0001",
    "status": "pending",
    "priority": "normal",
    "customer": {
      "id": 1,
      "name": "John Doe"
    },
    "vehicle": {
      "id": 1,
      "plate_number": "ABC-1234"
    },
    "description": "Tire replacement and alignment",
    "estimated_hours": 2.0,
    "created_at": "2026-07-04T12:00:00Z"
  }
}
```

### Update Service Job

**Endpoint:** `PUT /api/v1/service-jobs/:id`

**Permissions:** `service:update`

**Request:**
```json
{
  "status": "in_progress",
  "diagnosis": "Front tires worn, alignment off",
  "started_at": "2026-07-04T13:00:00Z"
}
```

**Response (200 OK):**
```json
{
  "data": {
    "id": 1,
    "job_number": "JOB-2026-0001",
    "status": "in_progress",
    "diagnosis": "Front tires worn, alignment off",
    "started_at": "2026-07-04T13:00:00Z",
    "updated_at": "2026-07-04T13:00:00Z"
  }
}
```

### Add Item to Service Job

**Endpoint:** `POST /api/v1/service-jobs/:id/items`

**Permissions:** `service:update`

**Request:**
```json
{
  "product_id": 1,
  "quantity": 2,
  "unit_price": 120.00
}
```

**Response (201 Created):**
```json
{
  "data": {
    "id": 1,
    "service_job_id": 1,
    "product_id": 1,
    "description": "Michelin Primacy 4 205/55R16",
    "quantity": 2,
    "unit_price": 120.00,
    "total_price": 240.00,
    "created_at": "2026-07-04T13:30:00Z"
  }
}
```

### Complete Service Job

**Endpoint:** `POST /api/v1/service-jobs/:id/complete`

**Permissions:** `service:update`

**Request:**
```json
{
  "work_performed": "Replaced 2 front tires, performed wheel alignment",
  "actual_hours": 1.5,
  "notes": "Customer satisfied with work"
}
```

**Response (200 OK):**
```json
{
  "data": {
    "id": 1,
    "job_number": "JOB-2026-0001",
    "status": "completed",
    "work_performed": "Replaced 2 front tires, performed wheel alignment",
    "actual_hours": 1.5,
    "completed_at": "2026-07-04T14:30:00Z"
  }
}
```

### Delete Service Job

**Endpoint:** `DELETE /api/v1/service-jobs/:id`

**Permissions:** `service:delete`

**Response (204 No Content)**

## Invoices

### List Invoices

**Endpoint:** `GET /api/v1/invoices`

**Permissions:** `invoice:view`

**Query Parameters:**
- `status` - Filter by status (`draft`, `issued`, `paid`, `voided`)
- `payment_status` - Filter by payment status (`unpaid`, `partial`, `paid`, `refunded`, `voided`)
- `customer_id` - Filter by customer
- `invoice_number` - Filter by invoice number
- `created_at_gte` - Filter by date range
- `created_at_lte` - Filter by date range

**Response (200 OK):**
```json
{
  "data": [
    {
      "id": 1,
      "invoice_number": "INV-2026-0001",
      "status": "paid",
      "payment_status": "paid",
      "customer": {
        "id": 1,
        "name": "John Doe"
      },
      "vehicle": {
        "id": 1,
        "plate_number": "ABC-1234"
      },
      "service_job": {
        "id": 1,
        "job_number": "JOB-2026-0001"
      },
      "subtotal": 255.00,
      "tax_rate": 0.00,
      "tax_amount": 0.00,
      "discount": 0.00,
      "total_usd": 255.00,
      "exchange_rate": 4050.00,
      "total_khr": 1032750.00,
      "paid_amount": 255.00,
      "payment_method": "cash",
      "issued_at": "2026-07-01T14:30:00Z",
      "created_at": "2026-07-01T14:30:00Z"
    }
  ],
  "meta": {
    "page": 1,
    "per_page": 20,
    "total": 50,
    "total_pages": 3
  }
}
```

### Get Invoice

**Endpoint:** `GET /api/v1/invoices/:id`

**Permissions:** `invoice:view`

**Response (200 OK):**
```json
{
  "data": {
    "id": 1,
    "invoice_number": "INV-2026-0001",
    "status": "paid",
    "payment_status": "paid",
    "customer": {
      "id": 1,
      "name": "John Doe",
      "phone": "+855 12 345 678",
      "address": "Phnom Penh, Cambodia"
    },
    "vehicle": {
      "id": 1,
      "plate_number": "ABC-1234",
      "make": "Toyota",
      "model": "Camry"
    },
    "service_job": {
      "id": 1,
      "job_number": "JOB-2026-0001"
    },
    "items": [
      {
        "id": 1,
        "item_type": "product",
        "product_id": 1,
        "description": "Michelin Primacy 4 205/55R16",
        "quantity": 2,
        "unit_price_usd": 120.00,
        "total_usd": 240.00
      },
      {
        "id": 2,
        "item_type": "labor",
        "product_id": 10,
        "description": "Wheel Alignment",
        "quantity": 1,
        "unit_price_usd": 15.00,
        "total_usd": 15.00
      }
    ],
    "subtotal": 255.00,
    "tax_rate": 0.00,
    "tax_amount": 0.00,
    "discount": 0.00,
    "total_usd": 255.00,
    "exchange_rate": 4050.00,
    "total_khr": 1032750.00,
    "paid_amount": 255.00,
    "payment_method": "cash",
    "payment_notes": "Paid in full",
    "notes": "Thank you for your business!",
    "issued_at": "2026-07-01T14:30:00Z",
    "created_at": "2026-07-01T14:30:00Z"
  }
}
```

### Create Invoice (Walk-in Sale)

**Endpoint:** `POST /api/v1/invoices`

**Permissions:** `invoice:create`

**Request:**
```json
{
  "customer_id": 1,
  "vehicle_id": 1,
  "items": [
    {
      "product_id": 1,
      "quantity": 2,
      "unit_price_usd": 120.00
    },
    {
      "product_id": 10,
      "quantity": 1,
      "unit_price_usd": 15.00
    }
  ],
  "discount": 0.00,
  "exchange_rate": 4050.00,
  "notes": "Thank you for your business!",
  "payment_method": "cash",
  "payment_status": "paid"
}
```

**Response (201 Created):**
```json
{
  "data": {
    "id": 1,
    "invoice_number": "INV-2026-0001",
    "status": "issued",
    "payment_status": "paid",
    "customer": {
      "id": 1,
      "name": "John Doe"
    },
    "subtotal": 255.00,
    "total_usd": 255.00,
    "exchange_rate": 4050.00,
    "total_khr": 1032750.00,
    "created_at": "2026-07-04T12:00:00Z"
  }
}
```

### Create Invoice from Service Job

**Endpoint:** `POST /api/v1/service-jobs/:id/invoice`

**Permissions:** `invoice:create`

**Request:**
```json
{
  "discount": 5.00,
  "exchange_rate": 4050.00,
  "payment_method": "cash",
  "payment_status": "paid",
  "notes": "Thank you for your business!"
}
```

**Response (201 Created):**
```json
{
  "data": {
    "id": 1,
    "invoice_number": "INV-2026-0001",
    "status": "issued",
    "payment_status": "paid",
    "service_job_id": 1,
    "subtotal": 255.00,
    "discount": 5.00,
    "total_usd": 250.00,
    "exchange_rate": 4050.00,
    "total_khr": 1012500.00,
    "created_at": "2026-07-04T14:30:00Z"
  }
}
```

### Update Invoice

**Endpoint:** `PUT /api/v1/invoices/:id`

**Permissions:** `invoice:update`

**Request:**
```json
{
  "payment_status": "paid",
  "paid_amount": 255.00,
  "payment_method": "cash",
  "payment_notes": "Paid in full"
}
```

**Response (200 OK):**
```json
{
  "data": {
    "id": 1,
    "invoice_number": "INV-2026-0001",
    "payment_status": "paid",
    "paid_amount": 255.00,
    "payment_method": "cash",
    "updated_at": "2026-07-04T15:00:00Z"
  }
}
```

### Generate Invoice PDF

**Endpoint:** `GET /api/v1/invoices/:id/pdf`

**Permissions:** `invoice:view`

**Response:** PDF file (application/pdf)

**Headers:**
```
Content-Type: application/pdf
Content-Disposition: attachment; filename="INV-2026-0001.pdf"
```

### Void Invoice

**Endpoint:** `POST /api/v1/invoices/:id/void`

**Permissions:** `invoice:void`

**Request:**
```json
{
  "reason": "Customer requested cancellation"
}
```

**Response (200 OK):**
```json
{
  "data": {
    "id": 1,
    "invoice_number": "INV-2026-0001",
    "status": "voided",
    "payment_status": "voided",
    "voided_at": "2026-07-04T15:30:00Z",
    "void_reason": "Customer requested cancellation"
  }
}
```

## Settings

### Get All Settings

**Endpoint:** `GET /api/v1/settings`

**Permissions:** `settings:view`

**Response (200 OK):**
```json
{
  "data": {
    "exchange_rate_usd_khr": 4050.00,
    "tax_rate_percent": 0.00,
    "tax_enabled": false,
    "invoice_prefix": "INV",
    "low_stock_threshold": 5,
    "telegram_enabled": false,
    "telegram_bot_token": "",
    "telegram_chat_id": ""
  }
}
```

### Update Settings

**Endpoint:** `PUT /api/v1/settings`

**Permissions:** `settings:update`

**Request:**
```json
{
  "exchange_rate_usd_khr": 4100.00,
  "tax_enabled": true,
  "tax_rate_percent": 10.00
}
```

**Response (200 OK):**
```json
{
  "data": {
    "exchange_rate_usd_khr": 4100.00,
    "tax_rate_percent": 10.00,
    "tax_enabled": true
  }
}
```

### Get Exchange Rate

**Endpoint:** `GET /api/v1/settings/exchange-rate`

**Permissions:** Public (or `settings:view`)

**Response (200 OK):**
```json
{
  "data": {
    "rate": 4050.00,
    "updated_at": "2026-07-01T10:00:00Z"
  }
}
```

### Update Exchange Rate

**Endpoint:** `PUT /api/v1/settings/exchange-rate`

**Permissions:** `settings:update`

**Request:**
```json
{
  "rate": 4100.00
}
```

**Response (200 OK):**
```json
{
  "data": {
    "rate": 4100.00,
    "updated_at": "2026-07-04T12:00:00Z"
  }
}
```

## Dashboard

### Get Dashboard Summary

**Endpoint:** `GET /api/v1/dashboard/summary`

**Permissions:** `report:view`

**Response (200 OK):**
```json
{
  "data": {
    "today": {
      "date": "2026-07-04",
      "revenue_usd": 1250.00,
      "revenue_khr": 5062500.00,
      "jobs_completed": 8,
      "invoices_issued": 10,
      "new_customers": 2
    },
    "this_week": {
      "revenue_usd": 8500.00,
      "jobs_completed": 45,
      "invoices_issued": 52
    },
    "this_month": {
      "revenue_usd": 32000.00,
      "jobs_completed": 180,
      "invoices_issued": 210
    },
    "low_stock_count": 5,
    "pending_jobs": 3
  }
}
```

### Get Daily Revenue

**Endpoint:** `GET /api/v1/dashboard/daily-revenue`

**Permissions:** `report:view`

**Query Parameters:**
- `days` - Number of days to look back (default: 30)

**Response (200 OK):**
```json
{
  "data": [
    {
      "date": "2026-07-04",
      "revenue_usd": 1250.00,
      "invoice_count": 10
    },
    {
      "date": "2026-07-03",
      "revenue_usd": 980.00,
      "invoice_count": 8
    },
    {
      "date": "2026-07-02",
      "revenue_usd": 1100.00,
      "invoice_count": 9
    }
  ]
}
```

### Get Low Stock Alerts

**Endpoint:** `GET /api/v1/dashboard/low-stock`

**Permissions:** `inventory:view`

**Response (200 OK):**
```json
{
  "data": [
    {
      "id": 5,
      "sku": "TIRE-BRI-19565R15",
      "name": "Bridgestone Turanza 195/65R15",
      "stock_quantity": 3,
      "min_stock_alert": 5,
      "sell_price": 95.00
    },
    {
      "id": 12,
      "sku": "PART-OIL-FILTER-001",
      "name": "Oil Filter Toyota",
      "stock_quantity": 2,
      "min_stock_alert": 10,
      "sell_price": 8.00
    }
  ]
}
```

## Telegram Bot (Optional)

### Send Invoice PDF to Telegram

**Endpoint:** `POST /api/v1/telegram/send-invoice`

**Permissions:** `invoice:view`

**Request:**
```json
{
  "invoice_id": 1,
  "chat_id": "-1001234567890"
}
```

**Response (200 OK):**
```json
{
  "data": {
    "message_id": 123,
    "sent_at": "2026-07-04T12:00:00Z"
  }
}
```

**Error (503 Service Unavailable):**
```json
{
  "error": {
    "code": "TELEGRAM_DISABLED",
    "message": "Telegram bot is not enabled"
  }
}
```

## Rate Limiting

API endpoints are rate-limited to prevent abuse:
- **Login**: 5 requests per minute
- **General endpoints**: 100 requests per minute per user
- **PDF generation**: 10 requests per minute

**Rate Limit Headers:**
```
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 95
X-RateLimit-Reset: 1625412345
```

## Error Handling

All errors follow a consistent format:

```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Request validation failed",
    "details": {
      "field": "email",
      "message": "Invalid email format"
    }
  }
}
```

### Common HTTP Status Codes
- `200 OK` - Success
- `201 Created` - Resource created
- `204 No Content` - Success with no response body
- `400 Bad Request` - Invalid request
- `401 Unauthorized` - Authentication required
- `403 Forbidden` - Insufficient permissions
- `404 Not Found` - Resource not found
- `409 Conflict` - Resource already exists
- `422 Unprocessable Entity` - Validation error
- `429 Too Many Requests` - Rate limit exceeded
- `500 Internal Server Error` - Server error
- `503 Service Unavailable` - Service temporarily unavailable

## API Versioning

The API uses URL-based versioning:
- Current version: `/api/v1`
- Future versions: `/api/v2`, `/api/v3`, etc.

When breaking changes are introduced, a new version will be created. Old versions will be supported for at least 6 months.
