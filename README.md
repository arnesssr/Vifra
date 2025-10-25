<div align="center">
  <img src="assets/vifra.png" alt="Vifra - VPS Infrastructure Monitor" width="300" height="300">
  
  # Vifra - VPS Infrastructure Monitor
  
  A comprehensive monitoring solution for VPS servers with real-time metrics, alerting, and multi-server management.
</div>

## Project Status

This project is currently in active development. The backend API is partially implemented with the following features:

- [x] Project structure and Go module setup
- [x] Database schema design with GORM models
- [x] REST API endpoints for servers, metrics, and alerts
- [x] Authentication system with JWT
- [x] Database migrations
- [x] Docker configuration for easy deployment
- [ ] Complete API implementation
- [ ] Frontend dashboard
- [ ] Agent for data collection
- [ ] Comprehensive test suite
- [ ] Documentation

## Features

### Current Features
- RESTful API for server management
- User authentication with JWT
- PostgreSQL database with GORM ORM
- Dockerized deployment
- Basic server CRUD operations

### Planned Features
- Real-time metrics collection
- Alerting system with notifications
- Multi-server dashboard
- Historical data visualization
- Lightweight agent for data collection
- Comprehensive test suite

## Technology Stack

- **Backend**: Go with Gorilla Mux and GORM
- **Database**: PostgreSQL
- **Authentication**: JWT
- **Deployment**: Docker & Docker Compose
- **Frontend**: React/Vue.js (planned)

## Quick Start

1. **Clone the repository**:
   ```bash
   git clone https://github.com/username/vps-monitor.git
   cd vps-monitor
   ```

2. **Set up environment variables**:
   ```bash
   cp configs/.env.example .env
   # Edit .env with your configuration
   ```

3. **Run with Docker Compose**:
   ```bash
   docker-compose up -d
   ```

4. **Run database migrations**:
   ```bash
   go run cmd/migrate/main.go
   ```

5. **Run the server**:
   ```bash
   go run cmd/server/main.go
   ```

## API Documentation

### Authentication
- `POST /api/v1/auth/login` - User login
- `POST /api/v1/auth/logout` - User logout

### Servers
- `GET /api/v1/servers` - List all monitored servers
- `GET /api/v1/servers/{id}` - Get server details
- `POST /api/v1/servers` - Add new server
- `PUT /api/v1/servers/{id}` - Update server
- `DELETE /api/v1/servers/{id}` - Remove server

### Metrics
- `GET /api/v1/servers/{id}/metrics` - Get real-time metrics
- `GET /api/v1/servers/{id}/metrics/history` - Get historical metrics

### Alerts
- `GET /api/v1/alerts` - List all alerts
- `POST /api/v1/alerts` - Create new alert rule
- `PUT /api/v1/alerts/{id}` - Update alert rule
- `DELETE /api/v1/alerts/{id}` - Delete alert rule

## Contributing

We welcome contributions from the community! This project is designed to be beginner-friendly with clear issues and good documentation.

### How to Contribute

1. Fork the repository
2. Create a feature branch
3. Commit your changes
4. Push to the branch
5. Create a Pull Request

### Good First Issues

Look for issues tagged with `good first issue` for beginner-friendly tasks.

## License

MIT

## Contact

For questions or feedback, please open an issue on GitHub.