# Phase 1: Initial Setup & Authentication

## Goals
Establish the foundation of the ERP Digital Printing backend with modular architecture and secure authentication.

## Completed Tasks

### 1. Project Initialization
- [x] Setup Fiber v3 with custom configuration.
- [x] Integrated GORM with PostgreSQL.
- [x] Structured logging using Zap and Lumberjack.
- [x] Configured environment management with Viper.

### 2. Shared Packages & Core Logic
- [x] Implemented standardized API response using Go Generics and `any`.
- [x] Created JWT utility for token generation and validation.
- [x] Setup Dependency Injection container for modularity.

### 3. User Module
- [x] Domain model with UUID support.
- [x] Repository implementation using GORM.
- [x] Usecase for business logic.
- [x] HTTP Handlers for CRUD operations.

### 4. Authentication Module
- [x] Login functionality.
- [x] Hybrid JWT strategy:
    - **Access Token**: Returned in JSON body.
    - **Refresh Token**: Stored in HttpOnly, Secure, and SameSite cookie.
- [x] Auth Middleware for protected routes.

## Technical Details
- **Architecture**: Modular Clean Architecture.
- **Language**: Go 1.22+.
- **Web Framework**: Fiber v3.
- **ORM**: GORM.
- **Database**: PostgreSQL.
