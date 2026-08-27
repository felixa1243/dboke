# Dboke - Universal Database Manager

Dboke is a highly secure, modern web-based database management tool built with a **Golang** backend (Clean Architecture) and a **Next.js** frontend (App Router, Tailwind v4).

## Monorepo Structure

- `/apps/api` - The Golang core backend (API Gateway, DB Adapters, Security Middleware).
- `/apps/web` - The Next.js frontend (Glassmorphism UI, React Components).
- `/packages` - Shared code, types, or configurations.

## Key Features

### 🛡️ Enterprise-Grade Security
- **AES-256-GCM Encryption:** Database credentials are symmetrically encrypted in-memory and bound directly to user sessions. No plaintext passwords are ever stored.
- **Zero-Trust Sessions:** Cryptographically secure 32-byte session IDs via HTTP-Only, SameSite=Lax cookies.
- **CSRF Protection:** Strict `X-CSRF-Token` header validation for all state-changing API endpoints.

### 🏗️ Architecture & Backend
- **Golang Clean Architecture:** Strict separation of concerns (Delivery, Services, Domain, Infrastructure).
- **Dynamic Connection Pooling:** Isolated `pgx/v5` connection pools spawned dynamically per database context.
- **Live Schema Introspection:** Real-time extraction of Postgres database schemas, including dynamic metric aggregation for precise row counts and data footprint sizes.

### 🎨 Frontend Excellence
- **Next.js 14 App Router:** Highly optimized, SEO-friendly React framework.
- **Tailwind CSS v4 & Zustand:** Premium glassmorphism UI, robust local state management, and seamless Light/Dark mode toggling.
- **Responsive Layout:** Sidebar-driven navigation with dynamic routing and micro-interactions.

## Prerequisites

- **Go** 1.21+
- **Node.js** 20+
- **npm** or **pnpm**

## Getting Started

### 1. Environment Setup

Copy the example environment file to the root of the project:
```bash
cp .env.example .env
```

Generate a secure master key for AES-256 database credential encryption:
```bash
cd apps/api
go run cmd/keygen/main.go
```
*Copy the generated 32-character key and paste it as `DBOKE_MASTER_KEY` in your `.env` file.*

### 2. Install Dependencies

Install Node.js dependencies for the workspace (from the root directory):
```bash
npm install
```

Install Golang dependencies:
```bash
cd apps/api
go mod tidy
```

### 3. Running the Project Locally

You will need two terminal windows to run the full stack locally.

**Terminal 1: Start the Golang API**
```bash
cd apps/api
go run cmd/api/main.go
```
*The API will start securely on `http://localhost:8080`.*

**Terminal 2: Start the Next.js Web App**
```bash
cd apps/web
npm run dev
```
*The frontend will start on `http://localhost:3000`. Open this in your browser to view the login interface.*

