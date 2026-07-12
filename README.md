# AutoStock - Garage Management System

A comprehensive web-based garage management system designed for small automotive workshops in Cambodia, specializing in tire sales and general auto repair services.

## Quick Links

- [Documentation](./docs/README.md) - Complete project documentation
- [Architecture](./docs/ARCHITECTURE.md) - System design and technical decisions
- [Database Schema](./docs/DATABASE.md) - Data models and relationships
- [API Documentation](./docs/API.md) - REST API endpoints and specifications
- [Frontend Guide](./docs/FRONTEND.md) - React application structure
- [Deployment Guide](./docs/DEPLOYMENT.md) - Production deployment instructions
- [Development Guide](./docs/DEVELOPMENT.md) - Development workflow and standards
- [Roadmap](./docs/ROADMAP.md) - Project phases and milestones

## Overview

AutoStock helps small garages manage their daily operations:
- **Inventory Management**: Track tires (with detailed specs), parts, and labor
- **Customer Management**: Customer profiles with vehicle information and service history
- **Service Jobs**: Track repair jobs from creation to completion
- **Invoice Generation**: Create invoices for service jobs or walk-in sales, generate PDFs
- **Dual Currency**: USD as primary currency, KHR as secondary with configurable exchange rates
- **Dashboard**: Daily revenue, completed jobs, low stock alerts

## Tech Stack

- **Backend**: Go + Gin + PostgreSQL
- **Frontend**: React + TypeScript + Vite + TanStack Query
- **Deployment**: Docker Compose
- **Styling**: Tailwind CSS

## Quick Start

```bash
# Clone the repository
git clone <repository-url>
cd AutoStock

# Copy environment file
cp .env.example .env

# Edit .env with your settings
nano .env

# Start all services
docker-compose up -d

# Access the application
# Frontend: http://localhost:3000
# Backend API: http://localhost:8080
```

## Features

### MVP Features
- ✅ User management (admin/staff roles)
- ✅ Inventory management (tires with specs, parts, labor)
- ✅ Customer & vehicle tracking
- ✅ Service job management
- ✅ Invoice generation (walk-in + service-based)
- ✅ Dual currency (USD/KHR)
- ✅ Printable receipts
- ✅ Dashboard with metrics
- ✅ Settings management
- 🔲 Telegram integration (placeholder)

### Future Features
- Khmer language support
- Multi-branch support
- Tax calculation module
- Payment integration
- Offline PWA
- Advanced reporting

## Project Structure

```
AutoStock/
├── backend/           # Go backend application
├── frontend/          # React frontend application
├── docs/              # Complete documentation
├── docker-compose.yml # Docker orchestration
├── .env.example       # Environment variables template
└── README.md         # This file
```

## Documentation

See the [docs](./docs/) directory for comprehensive documentation:

- **Getting Started**: [docs/README.md](./docs/README.md)
- **Architecture**: [docs/ARCHITECTURE.md](./docs/ARCHITECTURE.md)
- **Database**: [docs/DATABASE.md](./docs/DATABASE.md)
- **API**: [docs/API.md](./docs/API.md)
- **Frontend**: [docs/FRONTEND.md](./docs/FRONTEND.md)
- **Deployment**: [docs/DEPLOYMENT.md](./docs/DEPLOYMENT.md)
- **Development**: [docs/DEVELOPMENT.md](./docs/DEVELOPMENT.md)
- **Roadmap**: [docs/ROADMAP.md](./docs/ROADMAP.md)

## Development

See [docs/DEVELOPMENT.md](./docs/DEVELOPMENT.md) for detailed development instructions.

### Quick Development Setup

```bash
# Backend
cd backend
go run ./cmd/server

# Frontend
cd frontend
npm run dev
```

## Deployment

See [docs/DEPLOYMENT.md](./docs/DEPLOYMENT.md) for deployment instructions.

### Quick Deploy

```bash
# Using Docker Compose
docker-compose up -d

# Access at http://localhost:3000
```

## Support

For issues and questions, please open an issue in the repository.

## License

Proprietary - All rights reserved
