# Development Guide

## Overview

This guide covers the development workflow, coding standards, and best practices for AutoStock.

## Prerequisites

### Backend (Go)
- Go 1.21+
- PostgreSQL 15+
- SQLC (for type-safe SQL)
- golang-migrate (for database migrations)

### Frontend (React)
- Node.js 18+
- npm 9+

### Tools
- Docker & Docker Compose
- Git
- VS Code (recommended) or your preferred IDE

## Local Development Setup

### 1. Clone Repository

```bash
git clone https://github.com/your-org/autostock.git
cd autostock
```

### 2. Backend Setup

```bash
cd backend

# Install dependencies
go mod download

# Install development tools
go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# Copy environment file
cp .env.example .env

# Edit .env with your settings
nano .env
```

**Backend .env:**
```bash
DATABASE_URL=postgres://autostock:password@localhost:5432/autostock?sslmode=disable
JWT_SECRET=your-dev-secret-min-32-chars
PORT=8080
LOG_LEVEL=debug
TELEGRAM_BOT_TOKEN=
TELEGRAM_CHAT_ID=
```

### 3. Frontend Setup

```bash
cd frontend

# Install dependencies
npm install

# Initialize shadcn/ui (first time only)
npx shadcn-ui@latest init

# Add required components
npx shadcn-ui@latest add button input select dialog table card badge toast label

# Copy environment file
cp .env.example .env

# Edit .env
nano .env
```

**Frontend .env:**
```bash
VITE_API_URL=http://localhost:8080/api/v1
```

**Adding More shadcn/ui Components:**
```bash
# Add components as needed
npx shadcn-ui@latest add calendar
npx shadcn-ui@latest add dropdown-menu
npx shadcn-ui@latest add tabs
```

### 4. Database Setup

```bash
# Start PostgreSQL (using Docker)
docker-compose up -d postgres

# Or use local PostgreSQL
createdb autostock

# Run migrations
cd backend
migrate -path ./migrations -database "$DATABASE_URL" up

# Seed initial data (optional)
psql $DATABASE_URL -f ./migrations/seed.sql
```

### 5. Start Development Servers

**Backend:**
```bash
cd backend
go run ./cmd/server
```

**Frontend:**
```bash
cd frontend
npm run dev
```

**Access:**
- Frontend: http://localhost:3000
- Backend API: http://localhost:8080
- API Docs: http://localhost:8080/swagger

## Development Workflow

### 1. Create Feature Branch

```bash
git checkout -b feature/your-feature-name
```

### 2. Make Changes

- Write code following the coding standards below
- Write tests for new functionality
- Update documentation if needed

### 3. Run Tests

```bash
# Backend
cd backend
go test ./...

# Frontend
cd frontend
npm run test
```

### 4. Run Linters

```bash
# Backend
cd backend
golangci-lint run

# Frontend
cd frontend
npm run lint
```

### 5. Commit Changes

```bash
git add .
git commit -m "feat: add your feature description"
```

**Commit Message Format:**
```
<type>: <description>

[optional body]

[optional footer]
```

**Types:**
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `style`: Code style changes (formatting, etc.)
- `refactor`: Code refactoring
- `test`: Adding or updating tests
- `chore`: Maintenance tasks

**Examples:**
```
feat: add customer search functionality

fix: resolve invoice calculation error

docs: update API documentation for invoices

refactor: simplify product repository queries
```

### 6. Push and Create Pull Request

```bash
git push origin feature/your-feature-name
```

Then create a PR on GitHub.

## Backend Development

### Project Structure

```
backend/
├── cmd/server/          # Main entry point
├── internal/            # Private application code
│   ├── auth/           # Authentication module
│   ├── config/         # Configuration
│   ├── customer/       # Customer module
│   ├── database/       # Database connection
│   ├── domain/         # Shared types
│   ├── inventory/      # Inventory module
│   ├── invoice/        # Invoice module
│   ├── middleware/     # HTTP middleware
│   ├── service/        # Service jobs module
│   ├── settings/       # Settings module
│   └── telegram/       # Telegram bot
├── migrations/         # Database migrations
├── pkg/                # Reusable packages
└── sqlc/               # SQLC generated code
```

### Module Structure

Each module follows this structure:

```
module/
├── handler/            # HTTP handlers
│   └── handler.go
├── service/            # Business logic
│   └── service.go
├── repository/         # Data access
│   ├── repository.go
│   └── queries.sql
├── models/             # Domain models
│   └── models.go
└── dto/                # Request/Response DTOs
    └── dto.go
```

### Writing Handlers

```go
// internal/inventory/handler/handler.go
package handler

import (
    "net/http"
    "strconv"
    
    "github.com/gin-gonic/gin"
    "autostock/internal/inventory/service"
    "autostock/internal/inventory/dto"
)

type Handler struct {
    service *service.Service
}

func NewHandler(service *service.Service) *Handler {
    return &Handler{service: service}
}

// ListProducts handles GET /api/v1/products
func (h *Handler) ListProducts(c *gin.Context) {
    // Parse query parameters
    filter := dto.ProductFilter{
        Type:     c.Query("type"),
        Page:     parseIntParam(c, "page", 1),
        PerPage:  parseIntParam(c, "per_page", 20),
    }
    
    // Call service
    products, total, err := h.service.ListProducts(c.Request.Context(), filter)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{
            "error": gin.H{
                "code": "INTERNAL_ERROR",
                "message": "Failed to list products",
            },
        })
        return
    }
    
    // Return response
    c.JSON(http.StatusOK, gin.H{
        "data": products,
        "meta": gin.H{
            "page": filter.Page,
            "per_page": filter.PerPage,
            "total": total,
        },
    })
}

// CreateProduct handles POST /api/v1/products
func (h *Handler) CreateProduct(c *gin.Context) {
    var req dto.CreateProductRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{
            "error": gin.H{
                "code": "VALIDATION_ERROR",
                "message": err.Error(),
            },
        })
        return
    }
    
    product, err := h.service.CreateProduct(c.Request.Context(), &req)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{
            "error": gin.H{
                "code": "INTERNAL_ERROR",
                "message": "Failed to create product",
            },
        })
        return
    }
    
    c.JSON(http.StatusCreated, gin.H{
        "data": product,
    })
}

func parseIntParam(c *gin.Context, key string, defaultValue int) int {
    value := c.Query(key)
    if value == "" {
        return defaultValue
    }
    
    num, err := strconv.Atoi(value)
    if err != nil {
        return defaultValue
    }
    
    return num
}
```

### Writing Services

```go
// internal/inventory/service/service.go
package service

import (
    "context"
    "autostock/internal/inventory/repository"
    "autostock/internal/inventory/dto"
    "autostock/internal/domain"
)

type Service struct {
    repo *repository.Repository
}

func NewService(repo *repository.Repository) *Service {
    return &Service{repo: repo}
}

func (s *Service) ListProducts(ctx context.Context, filter dto.ProductFilter) ([]*domain.Product, int64, error) {
    // Business logic here
    products, err := s.repo.List(ctx, repository.ProductFilter{
        Type:     filter.Type,
        Limit:    filter.PerPage,
        Offset:   (filter.Page - 1) * filter.PerPage,
    })
    if err != nil {
        return nil, 0, err
    }
    
    total, err := s.repo.Count(ctx, repository.ProductFilter{
        Type: filter.Type,
    })
    if err != nil {
        return nil, 0, err
    }
    
    return products, total, nil
}

func (s *Service) CreateProduct(ctx context.Context, req *dto.CreateProductRequest) (*domain.Product, error) {
    // Validate business rules
    if req.SKU == "" {
        return nil, domain.ErrInvalidSKU
    }
    
    // Check if SKU already exists
    existing, err := s.repo.GetBySKU(ctx, req.SKU)
    if err == nil && existing != nil {
        return nil, domain.ErrDuplicateSKU
    }
    
    // Create product
    product := &domain.Product{
        Type:          req.Type,
        SKU:           req.SKU,
        Name:          req.Name,
        BuyPrice:      req.BuyPrice,
        SellPrice:     req.SellPrice,
        StockQuantity: req.StockQuantity,
    }
    
    if err := s.repo.Create(ctx, product); err != nil {
        return nil, err
    }
    
    return product, nil
}
```

### Writing SQL Queries with SQLC

```sql
-- internal/inventory/repository/queries.sql
-- name: GetProduct :one
SELECT * FROM products
WHERE id = $1 AND is_active = true;

-- name: ListProducts :many
SELECT * FROM products
WHERE is_active = true
  AND ($1::text IS NULL OR type = $1)
  AND ($2::text IS NULL OR name ILIKE '%' || $2 || '%')
ORDER BY created_at DESC
LIMIT $3 OFFSET $4;

-- name: CountProducts :one
SELECT COUNT(*) FROM products
WHERE is_active = true
  AND ($1::text IS NULL OR type = $1);

-- name: CreateProduct :one
INSERT INTO products (
    branch_id, type, sku, name, description,
    buy_price, sell_price, stock_quantity, min_stock_alert,
    tire_size, tire_brand, tire_model
) VALUES (
    $1, $2, $3, $4, $5,
    $6, $7, $8, $9,
    $10, $11, $12
)
RETURNING *;

-- name: UpdateProduct :one
UPDATE products
SET name = COALESCE($1, name),
    sell_price = COALESCE($2, sell_price),
    stock_quantity = COALESCE($3, stock_quantity),
    updated_at = NOW()
WHERE id = $4
RETURNING *;

-- name: DeleteProduct :exec
UPDATE products
SET is_active = false, updated_at = NOW()
WHERE id = $1;
```

**Generate Go Code:**

```bash
cd backend
sqlc generate
```

### Writing Tests

```go
// internal/inventory/service/service_test.go
package service

import (
    "context"
    "testing"
    
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
    "autostock/internal/inventory/repository"
    "autostock/internal/inventory/dto"
    "autostock/internal/domain"
)

type MockRepository struct {
    mock.Mock
}

func (m *MockRepository) List(ctx context.Context, filter repository.ProductFilter) ([]*domain.Product, error) {
    args := m.Called(ctx, filter)
    return args.Get(0).([]*domain.Product), args.Error(1)
}

func TestListProducts(t *testing.T) {
    // Arrange
    mockRepo := new(MockRepository)
    service := NewService(mockRepo)
    
    expectedProducts := []*domain.Product{
        {ID: 1, Name: "Product 1"},
        {ID: 2, Name: "Product 2"},
    }
    
    mockRepo.On("List", mock.Anything, mock.Anything).Return(expectedProducts, nil)
    mockRepo.On("Count", mock.Anything, mock.Anything).Return(int64(2), nil)
    
    // Act
    products, total, err := service.ListProducts(context.Background(), dto.ProductFilter{
        Page:    1,
        PerPage: 20,
    })
    
    // Assert
    assert.NoError(t, err)
    assert.Equal(t, expectedProducts, products)
    assert.Equal(t, int64(2), total)
    mockRepo.AssertExpectations(t)
}
```

## Frontend Development

### Component Structure

```typescript
// components/inventory/ProductList.tsx
import { useProducts } from '../../hooks/useProducts';
import { Product } from '../../types';

export const ProductList: React.FC = () => {
  const { data, isLoading, error } = useProducts();
  
  if (isLoading) return <div>Loading...</div>;
  if (error) return <div>Error loading products</div>;
  
  return (
    <div className="space-y-4">
      <h1 className="text-2xl font-bold">Products</h1>
      
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
        {data?.data.map((product: Product) => (
          <ProductCard key={product.id} product={product} />
        ))}
      </div>
    </div>
  );
};
```

### Custom Hooks

```typescript
// hooks/useProducts.ts
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { productApi } from '../services/products';

export const useProducts = (filter?: ProductFilter) => {
  return useQuery({
    queryKey: ['products', filter],
    queryFn: () => productApi.list(filter),
  });
};

export const useCreateProduct = () => {
  const queryClient = useQueryClient();
  
  return useMutation({
    mutationFn: (data: CreateProductRequest) => productApi.create(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['products'] });
    },
  });
};
```

### Form Handling

```typescript
// components/inventory/ProductForm.tsx
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';

const productSchema = z.object({
  name: z.string().min(1, 'Name is required'),
  sku: z.string().min(1, 'SKU is required'),
  sell_price: z.number().min(0, 'Price must be positive'),
});

type ProductFormData = z.infer<typeof productSchema>;

export const ProductForm: React.FC<{ onSubmit: (data: ProductFormData) => void }> = ({ onSubmit }) => {
  const { register, handleSubmit, formState: { errors } } = useForm<ProductFormData>({
    resolver: zodResolver(productSchema),
  });
  
  return (
    <form onSubmit={handleSubmit(onSubmit)} className="space-y-2">
      <div className="space-y-1">
        <Label htmlFor="name">Product Name</Label>
        <Input id="name" {...register('name')} />
        {errors.name && <p className="text-xs text-destructive">{errors.name.message}</p>}
      </div>
      
      <div className="space-y-1">
        <Label htmlFor="sku">SKU</Label>
        <Input id="sku" {...register('sku')} />
        {errors.sku && <p className="text-xs text-destructive">{errors.sku.message}</p>}
      </div>
      
      <div className="space-y-1">
        <Label htmlFor="sell_price">Sell Price (USD)</Label>
        <Input id="sell_price" type="number" step="0.01" {...register('sell_price', { valueAsNumber: true })} />
        {errors.sell_price && <p className="text-xs text-destructive">{errors.sell_price.message}</p>}
      </div>
      
      <Button type="submit">Save Product</Button>
    </form>
  );
};
```

### shadcn/ui Workflow

**Adding a new component:**
```bash
# From frontend/ directory
npx shadcn-ui@latest add <component-name>

# Example
npx shadcn-ui@latest add dialog
```

**Component location:** `src/components/ui/<component>.tsx`

**Customizing a component:**
1. Open `src/components/ui/<component>.tsx`
2. Edit the component directly (you own the code)
3. Changes apply globally

**When to create a new component vs. customize existing:**
- **Customize existing:** When you need different styling/behavior for the same concept
- **Create new:** When you need a completely different component

**Mobile-first checklist:**
- [ ] Test on mobile viewport (320px width)
- [ ] Touch targets are 44x44px minimum
- [ ] Forms stack vertically on mobile
- [ ] Tables use card layout or horizontal scroll on mobile
- [ ] Modals use full-screen on mobile (Sheet component)

**Import path:** Use `@/components/ui/` alias
```tsx
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
```

## Coding Standards

### Go

**Formatting:**
- Use `gofmt` to format code
- Use `goimports` to manage imports

**Naming:**
- Package names: lowercase, single word
- Function names: CamelCase (exported) or camelCase (unexported)
- Variable names: camelCase
- Constants: CamelCase or UPPER_SNAKE_CASE

**Error Handling:**
```go
// Good
if err != nil {
    return fmt.Errorf("failed to create product: %w", err)
}

// Bad
if err != nil {
    log.Println(err)
    return err
}
```

**Context:**
```go
// Always pass context as first parameter
func (s *Service) CreateProduct(ctx context.Context, req *CreateRequest) (*Product, error) {
    // Use ctx for database calls
    product, err := s.repo.Create(ctx, req)
    if err != nil {
        return nil, err
    }
    return product, nil
}
```

### TypeScript / React

**Component Naming:**
- Use PascalCase for component names
- Use descriptive names (e.g., `ProductList` not `List`)

**Props:**
```typescript
// Good
interface ProductCardProps {
  product: Product;
  onEdit?: (id: number) => void;
}

export const ProductCard: React.FC<ProductCardProps> = ({ product, onEdit }) => {
  // ...
};

// Bad
export const ProductCard = (props: any) => {
  // ...
};
```

**Hooks:**
```typescript
// Good
const [count, setCount] = useState(0);

useEffect(() => {
  // Side effect
}, [dependency]);

// Bad
let count = 0; // Don't use let for state
```

**Async/Await:**
```typescript
// Good
const fetchProducts = async () => {
  try {
    const products = await productApi.list();
    setProducts(products);
  } catch (error) {
    console.error('Failed to fetch products:', error);
  }
};

// Bad
productApi.list().then(products => {
  setProducts(products);
});
```

## Database Migrations

### Creating Migrations

```bash
# Create new migration
migrate create -ext sql -dir ./migrations -seq add_tire_fields

# This creates:
# migrations/000002_add_tire_fields.up.sql
# migrations/000002_add_tire_fields.down.sql
```

**Up Migration:**
```sql
-- migrations/000002_add_tire_fields.up.sql
ALTER TABLE products
ADD COLUMN tire_size VARCHAR(50),
ADD COLUMN tire_brand VARCHAR(100),
ADD COLUMN tire_model VARCHAR(100);

CREATE INDEX idx_products_tire_size ON products(tire_size);
```

**Down Migration:**
```sql
-- migrations/000002_add_tire_fields.down.sql
DROP INDEX IF EXISTS idx_products_tire_size;

ALTER TABLE products
DROP COLUMN IF EXISTS tire_size,
DROP COLUMN IF EXISTS tire_brand,
DROP COLUMN IF EXISTS tire_model;
```

### Running Migrations

```bash
# Apply all migrations
migrate -path ./migrations -database "$DATABASE_URL" up

# Rollback last migration
migrate -path ./migrations -database "$DATABASE_URL" down 1

# Reset database
migrate -path ./migrations -database "$DATABASE_URL" reset
```

## Testing

### Backend Tests

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run specific test
go test -run TestCreateProduct ./internal/inventory/service

# Run tests with verbose output
go test -v ./...
```

### Frontend Tests

```bash
# Run all tests
npm run test

# Run tests in watch mode
npm run test:watch

# Run tests with coverage
npm run test:coverage

# Run specific test file
npm run test -- ProductList.test.tsx
```

## Debugging

### Backend Debugging

**Using Delve:**
```bash
# Install Delve
go install github.com/go-delve/delve/cmd/dlv@latest

# Debug test
dlv test ./internal/inventory/service

# Debug server
dlv debug ./cmd/server
```

**Logging:**
```go
import "github.com/rs/zerolog/log"

log.Debug().Str("sku", sku).Msg("Looking up product")
log.Error().Err(err).Int("id", id).Msg("Failed to fetch product")
```

### Frontend Debugging

**React DevTools:**
- Install React DevTools browser extension
- Inspect component tree
- View props and state

**Network Tab:**
- Check API requests in browser DevTools
- Verify request/response format
- Check for CORS errors

**Console Logging:**
```typescript
console.log('Product:', product);
console.error('Error:', error);
console.warn('Warning:', warning);
```

## Performance

### Backend

**Database Queries:**
- Use EXPLAIN ANALYZE to check query performance
- Add indexes for frequently queried columns
- Use connection pooling (already configured)

**Caching:**
```go
// Cache frequently accessed data
var settingsCache *Settings
var cacheMu sync.RWMutex

func (s *Service) GetSettings(ctx context.Context) (*Settings, error) {
    cacheMu.RLock()
    if settingsCache != nil {
        defer cacheMu.RUnlock()
        return settingsCache, nil
    }
    cacheMu.RUnlock()
    
    // Fetch from database
    settings, err := s.repo.GetSettings(ctx)
    if err != nil {
        return nil, err
    }
    
    cacheMu.Lock()
    settingsCache = settings
    cacheMu.Unlock()
    
    return settings, nil
}
```

### Frontend

**Memoization:**
```typescript
// Memoize expensive calculations
const filteredProducts = useMemo(() => {
  return products.filter(p => p.stock_quantity < p.min_stock_alert);
}, [products]);

// Memoize callbacks
const handleClick = useCallback((id: number) => {
  setSelectedId(id);
}, []);
```

**Code Splitting:**
```typescript
// Lazy load routes
const Dashboard = lazy(() => import('./pages/Dashboard'));
```

## Security

### Backend

**Input Validation:**
```go
// Validate all inputs
if req.SKU == "" {
    return nil, domain.ErrInvalidSKU
}

if req.SellPrice < 0 {
    return nil, domain.ErrInvalidPrice
}
```

**SQL Injection Prevention:**
- Use SQLC (parameterized queries)
- Never concatenate user input into SQL

**Authentication:**
```go
// Verify JWT token
func AuthMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        token := c.GetHeader("Authorization")
        if token == "" {
            c.AbortWithStatusJSON(401, gin.H{"error": "Unauthorized"})
            return
        }
        
        claims, err := jwt.ValidateToken(token)
        if err != nil {
            c.AbortWithStatusJSON(401, gin.H{"error": "Invalid token"})
            return
        }
        
        c.Set("user_id", claims.UserID)
        c.Next()
    }
}
```

### Frontend

**XSS Prevention:**
- React automatically escapes JSX
- Never use `dangerouslySetInnerHTML` with user input

**CSRF Protection:**
- Use SameSite cookies
- Verify Origin header for state-changing requests

## Git Workflow

### Branch Naming

```
feature/add-customer-search
fix/invoice-calculation-error
docs/update-api-documentation
refactor/simplify-product-queries
```

### Pull Request Process

1. Create feature branch from `main`
2. Make changes and commit
3. Push to remote
4. Create PR with description
5. Request review
6. Address feedback
7. Merge after approval

### Code Review Checklist

- [ ] Code follows style guide
- [ ] Tests are included
- [ ] Documentation is updated
- [ ] No security issues
- [ ] Performance is acceptable
- [ ] Error handling is proper

## Common Issues

### Port Already in Use

```bash
# Find process using port
lsof -i :8080

# Kill process
kill -9 <PID>
```

### Database Connection Issues

```bash
# Check if PostgreSQL is running
pg_isready

# Check connection string
echo $DATABASE_URL
```

### CORS Errors

```go
// Update CORS config in backend
corsConfig := cors.Config{
    AllowOrigins: []string{"http://localhost:3000"},
    AllowMethods: []string{"GET", "POST", "PUT", "DELETE"},
}
```

## Resources

### Go
- [Effective Go](https://go.dev/doc/effective_go)
- [Go by Example](https://gobyexample.com/)
- [SQLC Documentation](https://docs.sqlc.dev/)

### React
- [React Documentation](https://react.dev/)
- [TypeScript Handbook](https://www.typescriptlang.org/docs/)
- [TanStack Query](https://tanstack.com/query/latest)

### PostgreSQL
- [PostgreSQL Documentation](https://www.postgresql.org/docs/)
- [Useful PostgreSQL Queries](https://github.com/timescale/useful-queries)
