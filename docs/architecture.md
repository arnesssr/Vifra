# VPS Monitor System Architecture

## Overview

VPS Monitor is designed as a distributed monitoring system with two main components:
1. **Dashboard Server**: Central web application that collects, stores, and displays server metrics
2. **Agent**: Lightweight daemon that runs on each monitored server to collect metrics

## System Components

### Dashboard Server (Central Monitoring System)

The Dashboard Server is the core component that provides:

- **RESTful API**: For server management, metrics retrieval, and alerting
- **Database Storage**: PostgreSQL database for storing metrics, server information, and alerts
- **User Interface**: Web-based dashboard for visualization and management
- **Authentication**: JWT-based authentication and authorization
- **Alerting Engine**: Configurable alert rules and notification system

### Agent (Lightweight Daemon)

The Agent is a lightweight process that runs on each monitored VPS server:

- **Local Metrics Collection**: Collects CPU, memory, disk, and network metrics locally
- **Secure Communication**: Submits metrics to the Dashboard Server via HTTPS API
- **Agent Authentication**: Uses unique AgentKey for secure authentication
- **Minimal Resource Usage**: Designed to have minimal impact on server performance

## Deployment Architecture

### Distributed Model

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   VPS Server    │    │   VPS Server    │    │   VPS Server    │
│                 │    │                 │    │                 │
│  ┌───────────┐  │    │  ┌───────────┐  │    │  ┌───────────┐  │
│  │   Agent   │  │    │  │   Agent   │  │    │  │   Agent   │  │
│  │           │  │    │  │           │  │    │  │           │  │
│  │ Collects  │  │    │  │ Collects  │  │    │  │ Collects  │  │
│  │  Metrics  │  │    │  │  Metrics  │  │    │  │  Metrics  │  │
│  └─────┬─────┘  │    │  └─────┬─────┘  │    │  └─────┬─────┘  │
└────────┼────────┘    └────────┼────────┘    └────────┼────────┘
         │                      │                      │
         └──────────────────────┼──────────────────────┘
                                │
                     ┌──────────▼──────────┐
                     │  Dashboard Server   │
                     │                     │
                     │  ┌───────────────┐  │
                     │  │      API      │  │
                     │  ├───────────────┤  │
                     │  │   Database    │  │
                     │  ├───────────────┤  │
                     │  │ Alert Engine  │  │
                     │  ├───────────────┤  │
                     │  │   Dashboard   │  │
                     │  └───────────────┘  │
                     └─────────────────────┘
```

## Communication Flow

### 1. Setup Phase

1. Deploy Dashboard Server on a central location (VPS, cloud instance, or dedicated server)
2. Register VPS servers to monitor via the API or dashboard
3. Install lightweight agents on each VPS to be monitored
4. Agents receive unique AgentKey for authentication

### 2. Monitoring Phase

1. Agents collect metrics from their local VPS at regular intervals
2. Agents submit metrics to the Dashboard Server via HTTPS API
3. Dashboard Server validates AgentKey and stores metrics in PostgreSQL
4. Users access the dashboard to view real-time and historical data
5. Alerting system triggers notifications based on configured rules

### 3. Data Flow

```
[VPS Server] --(Local Metrics Collection)--> [Agent]
     │
     └── CPU Usage, Memory Usage, Disk Usage, Network I/O
     └── System Load, Process Information

[Agent] --(HTTPS API with AgentKey Auth)--> [Dashboard Server]
     │
     └── POST /api/v1/metrics
     └── Periodic submissions (e.g., every 30 seconds)

[Dashboard Server] --(Store in PostgreSQL)--> [Database]
     │
     └── Metrics data with timestamps
     └── Server information and metadata

[User] --(HTTPS)--> [Dashboard Server]
     │
     └── GET /api/v1/servers/{id}/metrics
     └── GET /api/v1/servers/{id}/metrics/history
     └── Web dashboard access
```

## Security Model

### Authentication Methods

1. **User Authentication**: JWT tokens for dashboard users
2. **Agent Authentication**: Unique AgentKey for each monitored server
3. **API Security**: HTTPS encryption for all communications

### Security Benefits

- **No SSH Required**: Eliminates need for SSH key management
- **Encrypted Communication**: All data transmitted over HTTPS
- **Agent Isolation**: Each agent has unique credentials
- **Rate Limiting**: Prevents abuse and DoS attacks
- **Audit Logging**: Tracks all security-relevant actions

## Advantages of This Architecture

### Security
- No need to store SSH keys or credentials
- Encrypted communication between all components
- AgentKey authentication prevents unauthorized access
- Reduced attack surface compared to SSH-based monitoring

### Scalability
- Can monitor hundreds or thousands of servers
- Agents operate independently
- Dashboard Server can be scaled horizontally
- Database can be optimized for time-series data

### Performance
- Agents collect data locally without network overhead
- Minimal resource usage on monitored servers
- Efficient data transmission and storage
- Real-time monitoring capabilities

### Reliability
- If one agent fails, others continue to work
- Dashboard Server provides high availability options
- Data persistence in PostgreSQL database
- Built-in retry mechanisms for failed submissions

## Deployment Scenarios

### Single Server Deployment
- Dashboard Server and database on one machine
- Suitable for monitoring a small number of servers
- Simple setup and maintenance

### Distributed Deployment
- Dashboard Server and database on separate machines
- Multiple Dashboard Servers for load balancing
- Database clustering for high availability
- Suitable for large-scale deployments

### Cloud Deployment
- Dashboard Server on cloud platform (AWS, GCP, Azure)
- Database as a service (RDS, Cloud SQL, etc.)
- Agents on various VPS providers
- Auto-scaling capabilities

## Network Requirements

### Inbound Connections
- Dashboard Server: HTTPS (typically port 443 or 8443)
- Database: PostgreSQL port (typically 5432)

### Outbound Connections
- Agents: HTTPS to Dashboard Server
- Dashboard Server: Database connections
- Optional: Notification services (email, webhook, etc.)

## Resource Requirements

### Dashboard Server
- CPU: 2+ cores recommended
- RAM: 4GB+ recommended
- Storage: Depends on number of servers and retention period
- Bandwidth: Based on number of agents and submission frequency

### Agent
- CPU: Minimal (typically < 1%)
- RAM: 10-50MB
- Storage: Minimal (no persistent storage required)
- Bandwidth: Low (small JSON payloads)

## Failure Handling

### Agent Failures
- Metrics collection pauses until agent restarts
- Dashboard shows server as offline after timeout
- No impact on other agents or Dashboard Server

### Network Issues
- Agents queue metrics during outages (with limits)
- Dashboard shows delayed data
- Automatic reconnection when network restored

### Dashboard Server Failures
- Agents continue collecting data locally
- Data may be lost if outage is prolonged
- High availability setup recommended for production