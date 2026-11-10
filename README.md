# StockFlow
A warehouse inventory management application built with Go, React, and PostgreSQL.
StockFlow helps manage product information, monitor inventory levels, track stock movements, and process sales orders.

## Tech Stack
- Backend: Go
- Frontend: React + TypeScript + Vite
- UI Library: Mantine
- Database: PostgreSQL
- Routing: Chi Router
- Containerization: Docker Compose

## Getting Started
### Prerequisites
Make sure the following tools are installed:
- Go
- Node.js and npm
- PostgreSQL
- Docker and Docker Compose (optional)

### 1. Clone the Repository
```bash
git clone
cd StockFlow
```

### 2. Database Setup
Create a PostgreSQL database for StockFlow.
Run the SQL migration files in order:
1. `backend/migrations/create_tables.sql`
2. `backend/migrations/seed_data.sql`

### 3. Backend Setup
Install dependencies:
```bash
go mod download
```

Run the backend server:
```bash
go run ./cmd/api
```

### 4. Frontend Setup
Install dependencies:
```bash
npm install
```

Run the development server:
```bash
npm run dev
```

## Development
### Backend
Build the backend:
```bash
go build ./...
```

### Frontend
Build the frontend:
```bash
npm run build
```