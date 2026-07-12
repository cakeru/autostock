# Architecture

## System Overview

AutoStock follows a **modular monolith** architecture. This approach was chosen over microservices because:
- The scale is small (5-10 customers/day, 1-3 users)
- Simpler deployment and maintenance
- Easier to develop and debug
- Can extract services later if needed

The system is divided into clear domain modules with well-defined boundaries, making it easy to evolve into microservices in the future if the business grows significantly.

## Architecture Diagram

```
┌─────────────────────────────────────────────────────────────┐
│                         Frontend                             │
│              React + TypeScript + shadcn/ui                  │
│           (Vite + TanStack Query + Tailwind)                 │
└────────────────────────┬────────────────────────────────────┘
                         │ HTTP/REST
                         ▼
┌─────────────────────────────────────────────────────────────┐
│                      Reverse Proxy                           │
│                    (Nginx / Traefik)                         │
└────────────────────────┬────────────────────────────────────┘
                         │
                         ▼
┌─────────────────────────────────────────────────────────────┐
│                      Backend API                             │
│                    Go + Gin Framework                        │
│  ┌──────────────────────────────────────────────────────┐  │
│  │                   Middleware Layer                    │  │
│  │  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌────────┐ │  │
│  │  │   Auth   │ │  CORS    │ │  Logger  │ │Permiss.│ │  │
│  │  └──────────┘ └──────────┘ └──────────┘ └────────┘ │  │
│  └──────────────────────────────────────────────────────┘  │
│                         │                                    │
│  ┌──────────────────────┴───────────────────────────────┐  │
│  │                  Domain Modules                       │  │
│  │  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌────────┐ │  │
│  │  │Inventory │ │ Customer │ │ Service  │ │Invoice │ │  │
│  │  └──────────┘ └──────────┘ └──────────┘ └────────┘ │  │
│  │  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌────────┐ │  │
│  │  │  Auth    │ │ Settings │ │Telegram  │ │Dashboard│ │  │
│  │  └──────────┘ └──────────┘ └──────────┘ └────────┘ │  │
│  └──────────────────────────────────────────────────────┘  │
│                         │                                    │
│  ┌──────────────────────┴───────────────────────────────┐  │
│  │                  Data Access Layer                    │  │
│  │              SQLC (Type-safe SQL)                     │  │
│  └──────────────────────────────────────────────────────┘  │
└────────────────────────┬────────────────────────────────────┘
                         │
                         ▼
┌─────────────────────────────────────────────────────────────┐
│                     PostgreSQL                               │
│                   (Primary Database)                         │
└─────────────────────────────────────────────────────────────┘
```

## Module Structure

Each domain module follows a consistent internal structure:

```
module/
├── handler/        # HTTP handlers (controllers)
│   └── handler.go
├── service/        # Business logic
│   └── service.go
├── repository/     # Data access (SQLC queries)
│   ├── repository.go
│   └── queries.sql
├── models/         # Domain models
│   └── models.go
└── dto/            # Data Transfer Objects (request/response)
    └── dto.go
```

### Module Responsibilities

#### Auth Module
- User authentication (JWT-based)
- Permission management
- Session handling
- Password hashing (bcrypt)

#### Inventory Module
- Product CRUD operations
- Tire-specific attributes management
- Stock level tracking
- Low stock alerts
- Product categorization (tires, parts, labor, consumables)

#### Customer Module
- Customer profile management
- Vehicle registration and tracking
- Service history
- Customer search and filtering

#### Service Module
- Service job creation and management
- Job status tracking (pending, in_progress, completed, cancelled)
- Linking jobs to customers, vehicles, and inventory
- Labor tracking

#### Invoice Module
- Invoice creation (service-based and walk-in sales)
- Invoice numbering (INV-YYYY-NNNN format)
- Dual currency calculation (USD/KHR)
- Tax calculation (optional)
- PDF generation
- Invoice voiding

#### Settings Module
- Garage configuration (name, logo, address, phone)
- Exchange rate management (global and per-invoice)
- System preferences
- Telegram bot configuration

#### Telegram Module (Optional)
- Bot initialization
- PDF invoice sending to configured channel/group
- Daily summary reports
- Notification management

#### Dashboard Module
- Daily revenue calculation
- Job completion statistics
- Low stock alerts
- Quick metrics aggregation

## Design Patterns

### 1. Repository Pattern
Data access is abstracted through repository interfaces, making it easy to:
- Swap database implementations
- Write unit tests with mock repositories
- Maintain clean separation of concerns

```go
type ProductRepository interface {
    Create(ctx context.Context, product *Product) error
    GetByID(ctx context.Context, id int64) (*Product, error)
    Update(ctx context.Context, product *Product) error
    Delete(ctx context.Context, id int64) error
    List(ctx context.Context, filter ProductFilter) ([]*Product, error)
}
```

### 2. Service Layer
Business logic resides in the service layer, which:
- Orchestrates multiple repositories
- Enforces business rules
- Handles transactions
- Remains independent of HTTP concerns

```go
type InvoiceService struct {
    invoiceRepo  InvoiceRepository
    productRepo  ProductRepository
    customerRepo CustomerRepository
    settingsRepo SettingsRepository
}

func (s *InvoiceService) CreateInvoice(ctx context.Context, req CreateInvoiceRequest) (*Invoice, error) {
    // Business logic here
}
```

### 3. Handler Layer
HTTP handlers:
- Parse and validate requests
- Call service methods
- Format and send responses
- Handle HTTP-specific concerns (status codes, headers)

```go
func (h *InvoiceHandler) CreateInvoice(c *gin.Context) {
    var req CreateInvoiceRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
        return
    }
    
    invoice, err := h.service.CreateInvoice(c.Request.Context(), req)
    if err != nil {
        c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
        return
    }
    
    c.JSON(http.StatusCreated, invoice)
}
```

### 4. Middleware Chain
HTTP requests pass through a chain of middleware:

```
Request → Logger → CORS → Auth → Permission → Handler → Response
```

## Multi-Tenancy Strategy

The system is designed for multi-branch support through **shared database with branch_id filtering**:

- All tenant-specific tables include a `branch_id` column
- Middleware extracts branch context from JWT token
- Repository queries automatically filter by `branch_id`
- No data leakage between branches

```go
// Middleware extracts branch context
func BranchMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        claims := GetJWTClaims(c)
        c.Set("branch_id", claims.BranchID)
        c.Next()
    }
}

// Repository queries filter by branch
func (r *productRepository) List(ctx context.Context, branchID int64) ([]*Product, error) {
    query := `SELECT * FROM products WHERE branch_id = $1`
    // ...
}
```

## Dual Currency Handling

All monetary values are stored in **USD** (primary currency). KHR amounts are calculated on-the-fly using exchange rates.

### Storage Strategy
```go
type Invoice struct {
    TotalUSD      float64 `json:"total_usd"`
    ExchangeRate  float64 `json:"exchange_rate"` // USD to KHR rate
    TotalKHR      float64 `json:"total_khr"`     // Calculated: TotalUSD * ExchangeRate
}
```

### Exchange Rate Sources
1. **Global Setting**: Default rate configured in settings (e.g., 4050 KHR = 1 USD)
2. **Per-Invoice Override**: Manual rate entry when creating invoice
3. **Historical Rate**: Stored with invoice for accurate historical reporting

### Currency Display
```go
func FormatCurrency(amount float64, currency string) string {
    if currency == "USD" {
        return fmt.Sprintf("$%.2f", amount)
    }
    return fmt.Sprintf("៛%.0f", amount) // KHR has no decimals
}
```

## PDF Generation

PDFs are generated using **chromedp** (headless Chrome) for maximum flexibility:

1. HTML template with CSS styling
2. Render HTML to PDF using chromedp
3. Return PDF as byte stream or save to file

### Template Structure
```html
<!DOCTYPE html>
<html>
<head>
    <style>
        /* Invoice styling */
    </style>
</head>
<body>
    <div class="invoice">
        <div class="header">
            <img src="{{.LogoURL}}" />
            <h1>{{.GarageName}}</h1>
            <p>{{.Address}}</p>
        </div>
        <div class="invoice-details">
            <p>Invoice #: {{.InvoiceNumber}}</p>
            <p>Date: {{.Date}}</p>
        </div>
        <!-- More sections -->
    </div>
</body>
</html>
```

## Security Considerations

### Authentication
- JWT tokens with configurable expiration
- Refresh token mechanism for long sessions
- Secure password hashing (bcrypt, cost 12)

### Authorization
- Role-based access control (Admin, Staff)
- Granular permission system
- Middleware-enforced permission checks

### Data Protection
- HTTPS in production (TLS termination at reverse proxy)
- SQL injection prevention (SQLC parameterized queries)
- XSS prevention (React's built-in escaping)
- CSRF protection (SameSite cookies)

### Audit Trail
- All critical actions logged to `audit_logs` table
- Tracks: user, action, entity, changes, timestamp

## Performance Considerations

### Database
- Indexed foreign keys
- Composite indexes for common queries
- Connection pooling (pgxpool)
- Query optimization with EXPLAIN ANALYZE

### Caching
- In-memory cache for frequently accessed data (settings, exchange rates)
- Cache invalidation on updates
- TTL-based expiration

### Frontend
- Code splitting with Vite
- Lazy loading of routes
- TanStack Query for data caching
- Optimistic updates for better UX

## Scalability Path

### Current State (MVP)
- Single VPS deployment
- PostgreSQL on same server
- 5-10 customers/day capacity

### Growth Path
1. **Vertical Scaling**: Upgrade VPS resources (CPU, RAM, storage)
2. **Database Scaling**: Move PostgreSQL to managed service (RDS, DigitalOcean)
3. **Horizontal Scaling**: Multiple backend instances behind load balancer
4. **CDN**: Serve static assets via CDN (CloudFront, Cloudflare)
5. **Microservices**: Extract high-load modules if needed

## Monitoring & Logging

### Logging
- Structured JSON logs with Zerolog
- Log levels: DEBUG, INFO, WARN, ERROR
- Request ID tracking across services
- Sensitive data redaction

### Metrics (Future)
- Prometheus metrics endpoint
- Grafana dashboards
- Alert rules for critical metrics

### Health Checks
- `/health` endpoint for liveness
- `/ready` endpoint for readiness
- Database connection status
- External service status (Telegram bot)

## Error Handling

### Backend
- Centralized error types
- Consistent error response format
- Error codes for frontend handling
- Stack traces in development, sanitized in production

```go
type AppError struct {
    Code    string `json:"code"`
    Message string `json:"message"`
    Status  int    `json:"-"`
}

// Response format
{
    "error": {
        "code": "PRODUCT_NOT_FOUND",
        "message": "Product with ID 123 not found"
    }
}
```

### Frontend
- Global error boundary
- Toast notifications for user feedback
- Form validation errors
- API error handling with retry logic

## Testing Strategy

### Unit Tests
- Service layer business logic
- Repository layer with mock database
- Utility functions

### Integration Tests
- API endpoint testing
- Database integration
- External service mocking (Telegram)

### End-to-End Tests (Future)
- Critical user flows
- Using Playwright or Cypress

### Test Coverage Target
- Minimum 70% coverage
- Focus on business-critical paths

## Future Enhancements

### Phase 2 (Post-MVP)
- **i18n**: Khmer language support with react-i18next
- **Multi-branch**: Enable branch management UI
- **Tax Module**: Configurable tax rates and calculations
- **Payment Integration**: Wing, ABA Pay, etc.
- **Offline PWA**: Service workers with sync queue

### Phase 3 (Advanced)
- **Mobile App**: React Native for mobile access
- **Advanced Reporting**: Custom report builder
- **API Webhooks**: Third-party integrations
- **AI Features**: Demand forecasting, automated reordering
