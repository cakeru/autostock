# AutoStock - Garage Management System

## Overview

AutoStock is a web-based garage management system designed for small automotive workshops in Cambodia, specializing in tire sales and general auto repair services. The system provides inventory management, customer tracking, service job management, and invoice generation with dual currency support (USD/KHR).

## Business Context

- **Target User**: Small garage/workshop in Cambodia
- **Primary Services**: Tire sales and changing, general auto repair
- **Scale**: 5-10 customers per day, 1-3 staff members
- **Location**: Single location (multi-branch ready for future)
- **Languages**: English (MVP), Khmer (post-MVP)

## Key Features

### MVP Features
- **User Management**: Admin and staff roles with modular permissions
- **Inventory Management**: Track tires (with detailed specs), parts, and labor
- **Customer Management**: Customer profiles with vehicle information and service history
- **Service Jobs**: Track repair jobs from creation to completion
- **Invoice Generation**: Create invoices for service jobs or walk-in sales, generate PDFs
- **Dual Currency**: USD as primary currency, KHR as secondary with configurable exchange rates
- **Dashboard**: Daily revenue, completed jobs, low stock alerts
- **Settings**: Garage branding, exchange rates, system configuration
- **Telegram Bot**: Optional integration for invoice notifications (configurable)

### Future Features (Post-MVP)
- Khmer language support (i18n)
- Multi-branch support
- Tax calculation module
- Payment gateway integration (Wing, ABA Pay, etc.)
- Consumables tracking
- Loyalty program
- Offline PWA with sync

## Tech Stack

### Backend
- **Language**: Go
- **Framework**: Gin
- **Database**: PostgreSQL
- **ORM**: SQLC (type-safe SQL code generation)
- **Logging**: Zerolog

### Frontend
- **Framework**: React + TypeScript
- **Build Tool**: Vite
- **State Management**: TanStack Query (React Query)
- **Form Handling**: React Hook Form + Zod
- **Styling**: Tailwind CSS
- **Component Library**: shadcn/ui (copy-paste components, not a dependency)
- **HTTP Client**: Axios

### Infrastructure
- **Containerization**: Docker
- **Orchestration**: Docker Compose
- **Deployment**: VPS or split (Vercel for frontend, VPS for backend)

## Quick Start

### Prerequisites
- Docker and Docker Compose
- Git

### Installation

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
# API Documentation: http://localhost:8080/swagger
```

### Default Credentials
- **Username**: admin
- **Password**: admin123 (change immediately after first login)

## Documentation

- [Architecture](./ARCHITECTURE.md) - System design and technical decisions
- [Database Schema](./DATABASE.md) - Data models and relationships
- [API Documentation](./API.md) - REST API endpoints and specifications
- [Frontend Guide](./FRONTEND.md) - React application structure and patterns
- [Deployment Guide](./DEPLOYMENT.md) - Production deployment instructions
- [Development Guide](./DEVELOPMENT.md) - Development workflow and standards
- [Roadmap](./ROADMAP.md) - Project phases and milestones

## Project Structure

```
AutoStock/
├── backend/                    # Go backend application
│   ├── cmd/server/            # Main entry point
│   ├── internal/              # Private application code
│   │   ├── auth/             # Authentication & permissions
│   │   ├── config/           # Configuration
│   │   ├── customer/         # Customer module
│   │   ├── database/         # Database connection
│   │   ├── domain/           # Shared types
│   │   ├── inventory/        # Inventory module
│   │   ├── invoice/          # Invoice module
│   │   ├── middleware/       # HTTP middleware
│   │   ├── service/          # Service jobs module
│   │   ├── settings/         # Settings module
│   │   └── telegram/         # Telegram bot (optional)
│   ├── migrations/           # Database migrations
│   ├── pkg/                  # Reusable packages
│   │   ├── currency/         # Dual currency helpers
│   │   ├── pdf/              # PDF generation
│   │   └── validator/        # Input validation
│   └── sqlc/                 # SQLC generated code
├── frontend/                  # React frontend application
│   ├── src/
│   │   ├── components/       # React components
│   │   ├── hooks/            # Custom hooks
│   │   ├── pages/            # Route pages
│   │   ├── services/         # API calls
│   │   ├── types/            # TypeScript types
│   │   └── utils/            # Utility functions
│   └── public/               # Static assets
├── docs/                      # Documentation
├── docker-compose.yml         # Docker orchestration
├── .env.example              # Environment variables template
└── README.md                 # This file
```

## Support

For issues and questions, please open an issue in the repository.

## License

Proprietary - All rights reserved
