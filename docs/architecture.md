# Vifra System Architecture

## Welcome to Vifra's Architecture

Vifra is a modern, agent-based VPS monitoring system built with security and scalability in mind. If you're evaluating Vifra for your infrastructure or considering contributing, this document will help you understand how everything works together.

## Core Philosophy

**Agent-Based Push Model**: Unlike traditional monitoring systems that SSH into your servers every minute, Vifra's agents actively push metrics to a central dashboard. This means:
- Your dashboard server doesn't need SSH access to monitored servers
- Reduced attack surface (no SSH key storage on the dashboard)
- Better scalability (agents work independently)
- Lower network overhead (efficient HTTPS API calls)

## System Components

## System Components

Vifra consists of two main components working together:

### 1. Dashboard Server (Your Control Center)

This is the central hub where all your monitoring data lives. You'll install this on a server you control (cloud VPS, dedicated server, or even your local machine for testing).

**What it does:**
- **RESTful API** - Receives metrics from agents, serves data to the web UI
- **PostgreSQL Database** - Stores all your metrics, server info, and alert history
- **Web Dashboard** - Clean UI to visualize your servers' health at a glance
- **Authentication** - JWT-based auth keeps your monitoring data secure
- **Alert Engine** - Real-time evaluation of metrics against your custom rules
- **Notification System** - Email, Slack, or webhooks when things go wrong

**Tech Stack:** Go (Gorilla Mux + GORM), PostgreSQL, React/Vue.js (frontend - coming soon)

### 2. Monitoring Agent (Runs on Each VPS)

Lightweight daemon that lives on every server you want to monitor. Think of it as your server's personal reporter.

**What it does:**
- **Collects Metrics Locally** - CPU, memory, disk, load average (< 1% CPU overhead)
- **Pushes to Dashboard** - HTTPS POST requests every 30 seconds (configurable)
- **Unique Authentication** - Each agent has its own AgentKey (no shared credentials)
- **Self-Contained** - ~10-50MB RAM, minimal disk usage, no dependencies

**Tech Stack:** Go (single binary), systemd integration on Linux

## How It All Connects

### The Big Picture

Here's how your Vifra deployment looks:

```
   Your Monitored Servers              Your Dashboard Server
 ┌─────────────────────────────┐     ┌───────────────────────┐
 │                             │     │                       │
 │  VPS 1: web-server-01      │     │   Vifra Dashboard     │
 │  ┌─────────────┐            │     │  ┌─────────────────┐  │
 │  │    Agent    │───────┐    │     │  │   REST API      │  │
 │  │ (Go Binary) │       │    │     │  ├─────────────────┤  │
 │  └─────────────┘       │    │     │  │  Alert Engine   │  │
 │                        │    │     │  ├─────────────────┤  │
 │  VPS 2: db-server-01   │    │     │  │  PostgreSQL DB  │  │
 │  ┌─────────────┐       │    │     │  ├─────────────────┤  │
 │  │    Agent    │───────┤    │     │  │  Web Dashboard  │  │
 │  │ (Go Binary) │       │    │     │  └─────────────────┘  │
 │  └─────────────┘       │    │     │                       │
 │                        │    │     └───────────────────────┘
 │  VPS 3: app-server-01  │    │              ▲
 │  ┌─────────────┐       │    │              │
 │  │    Agent    │───────┴────┼──────HTTPS───┘
 │  │ (Go Binary) │            │    (Port 443/8443)
 │  └─────────────┘            │
 │                             │
 └─────────────────────────────┘

   All agents push metrics to your dashboard via HTTPS API
   No inbound connections needed to monitored servers!
```

## How Communication Works

### Getting Started (One-Time Setup)

1. **Deploy your Dashboard Server**
   - Spin up a VPS/cloud instance or use a dedicated server
   - Run `docker-compose up` or build from source
   - Configure PostgreSQL connection

2. **Register your servers**
   - Add each VPS via the API: `POST /api/v1/servers`
   - You'll get back a unique `AgentKey` for each server
   - Note: You still need SSH to your servers for this initial setup!

3. **Install agents on your VPS servers**
   - SSH into each server (yes, you need SSH for this)
   - Run the install script: `sudo ./scripts/install-agent.sh`
   - Configure with your `AgentKey` and Dashboard URL
   - Enable the systemd service: `sudo systemctl enable vps-monitor-agent`

### Once Running (The Monitoring Loop)

Now here's where the "no SSH" magic happens:

**Every 30 seconds** (configurable):
```
1. Agent reads local system metrics
   │
   ├─ CPU: /proc/stat
   ├─ Memory: /proc/meminfo
   ├─ Disk: syscall.Statfs
   └─ Load: /proc/loadavg

2. Agent POSTs to dashboard
   │
   POST https://your-dashboard.com/api/v1/metrics
   Headers: Authorization: Bearer <AgentKey>
   Body: {
     "server_id": 1,
     "cpu_usage": 45.2,
     "memory_used": 4294967296,
     "memory_total": 8589934592,
     ...
   }

3. Dashboard validates & stores
   │
   ├─ Check AgentKey is valid
   ├─ Evaluate against alert rules
   ├─ Store in PostgreSQL
   └─ Send notifications if alerts triggered

4. You view the data
   │
   ├─ Web dashboard updates in real-time
   ├─ API endpoints serve historical data
   └─ Alerts arrive via email/Slack/webhook
```

**The key difference**: Dashboard never initiates connections to your servers. Agents always push outbound.

## Security: What You Need to Know

### Authentication & Encryption

**For Dashboard Users:**
- JWT tokens (login via username/password)
- Role-based access control (user vs admin)
- Secure session management

**For Agents:**
- Unique `AgentKey` per server (32-byte cryptographically random string)
- Each agent only has access to its own server's data
- Keys can be rotated if compromised

**For Communication:**
- All traffic over HTTPS/TLS
- Optional: Encrypt agent keys at rest in database
- Rate limiting on API endpoints (100 req/min per IP)
- Comprehensive audit logging

### The "No SSH" Reality Check

**What we mean by "No SSH required for monitoring":**
- ✅ Dashboard doesn't need SSH access to pull metrics
- ✅ No SSH keys stored on dashboard server
- ✅ No persistent SSH connections
- ✅ Reduced attack surface vs SSH-polling systems

**What you still need SSH for:**
- ⚠️ **Initial agent installation** (you have to deploy the binary somehow!)
- ⚠️ **Configuration changes** (editing `/etc/vps-monitor-agent/config.env`)
- ⚠️ **Agent updates** (deploying new versions)
- ⚠️ **Troubleshooting** (viewing logs, restarting service)
- ⚠️ **General server administration** (this is normal!)

**The advantage**: Dashboard server breach ≠ SSH access to all your servers. Agent keys only allow metric submission, not server control.

### Security Best Practices

1. **Run dashboard behind a reverse proxy** (nginx + Let's Encrypt)
2. **Use strong JWT secrets** (64+ random characters)
3. **Enable database encryption** (optional `ENCRYPTION_KEY` env var)
4. **Rotate agent keys periodically** (via API or dashboard)
5. **Monitor your monitor** (set up alerts for dashboard itself)
6. **Keep agents updated** (we'll add auto-update eventually)

## Why This Architecture?

### Security Benefits

**Compared to SSH-based monitoring tools:**

| Traditional (SSH polling) | Vifra (Agent push) |
|---------------------------|--------------------|
| Dashboard SSHs into servers every minute | Agents push outbound only |
| SSH keys on dashboard = single point of failure | Each agent has unique key |
| Dashboard compromise = all servers compromised | Dashboard compromise = metric visibility only |
| Firewall rules for inbound SSH | No inbound connections needed |

### Scalability Benefits

- **Horizontal scaling**: Add more dashboard servers behind a load balancer
- **Independent agents**: One server down doesn't affect others
- **Efficient data flow**: Small JSON payloads, minimal bandwidth
- **Database optimized**: Time-series data patterns, easy to shard

### Performance Benefits

- **Local collection**: No network latency for metric gathering
- **Low overhead**: < 1% CPU, < 50MB RAM per agent
- **Real-time**: Configurable intervals (10s to 5m typical)
- **Reliable**: Automatic retries with exponential backoff

### Practical Benefits

- **Easy deployment**: Single binary, systemd integration
- **Cloud-friendly**: Works across any VPS provider
- **Firewall-friendly**: Agents only need outbound HTTPS
- **Multi-tenant ready**: User isolation, role-based access

## Deployment Scenarios

### Starting Small (Recommended for Testing)

**Single-server setup:**
```
1 VPS running both Dashboard + PostgreSQL
  ├─ 2 CPU cores
  ├─ 4GB RAM
  ├─ 20GB SSD
  └─ Cost: ~$10-20/month

Good for: Testing, personal projects, monitoring up to 10-20 servers
```

**Quick start:**
```bash
git clone https://github.com/arnesssr/Vifra.git
cd Vifra
cp configs/.env.example .env  # Edit with your settings
docker-compose up -d
```

### Production (Recommended for Serious Use)

**Separate dashboard and database:**
```
Dashboard Server:           Database Server:
  ├─ 2-4 CPU cores            ├─ 4+ CPU cores
  ├─ 4-8GB RAM                ├─ 8-16GB RAM
  ├─ 50GB SSD                 ├─ 100GB+ SSD (depends on retention)
  └─ Load balancer ready      └─ Replication for HA

Good for: Production, monitoring 50+ servers, team use
```

### Enterprise (For Large Scale)

**Fully distributed:**
- Multiple dashboard servers behind load balancer
- PostgreSQL with read replicas
- Separate database for different metric types
- CDN for dashboard assets

Good for: 100s-1000s of servers, high availability requirements

## Network & Port Requirements

### What you need to open:

**On Dashboard Server:**
- **Inbound**: Port 443 (HTTPS) or 8080 (HTTP for testing)
- **Outbound**: Port 5432 to PostgreSQL, SMTP/webhooks for notifications

**On Monitored Servers (Agents):**
- **Inbound**: Nothing! (that's the point)
- **Outbound**: Port 443 to your dashboard server

**Firewall-friendly:** Agents only make outbound HTTPS calls, works through most corporate firewalls.

## Resource Requirements (Real Numbers)

### Dashboard Server

**Minimum (1-10 servers):**
- 1 CPU core
- 2GB RAM
- 10GB storage

**Recommended (10-50 servers):**
- 2 CPU cores
- 4GB RAM
- 50GB storage
- Retention: 30 days of metrics

**Production (50+ servers):**
- 4+ CPU cores
- 8GB+ RAM
- 100GB+ storage
- Retention: 90+ days

**Storage calculation:**
```
Per server: ~1KB per metric submission
30-second intervals = 120 submissions/hour
= 2,880 submissions/day
= ~3MB/day per server

10 servers × 30 days = ~900MB
50 servers × 90 days = ~13GB
```

### Agent (Per Monitored Server)

- **CPU**: < 0.5% average (barely noticeable)
- **RAM**: 15-50MB (lighter than most monitoring agents)
- **Disk**: < 10MB (single binary)
- **Network**: ~3KB/30s = ~250KB/hour

**Translation**: You won't notice the agent is running.

## What Happens When Things Break

### Agent Crashes or Stops

**What happens:**
- That server's metrics stop updating
- Dashboard shows "Offline" after 2 minutes
- Other servers keep working fine

**To fix:**
```bash
sudo systemctl restart vps-monitor-agent
sudo journalctl -u vps-monitor-agent -n 50  # Check logs
```

### Network Issues (Agent can't reach dashboard)

**What happens:**
- Agent keeps trying with exponential backoff
- Metrics queued in memory (up to 10 minutes worth)
- Dashboard shows stale data

**Recovery:**
- Automatic when network restored
- Queued metrics get submitted
- Gap in data if outage > 10 minutes

### Dashboard Server Down

**What happens:**
- Agents keep collecting locally
- Data gets queued (limited by agent memory)
- Metrics older than 10 minutes are lost

**Prevention:**
- Set up health checks
- Use load balancer with multiple dashboard instances
- Monitor your monitoring system!

### PostgreSQL Issues

**What happens:**
- API returns errors
- Agents keep retrying
- Recent data lost until database restored

**Prevention:**
- Regular backups (pg_dump)
- Replication for HA
- Connection pooling

## Contributing & Development

Interested in improving Vifra? Check out:
- [PRD.md](PRD.md) - Product requirements and roadmap
- [agent.md](agent.md) - Agent installation and configuration
- [Contributing Guide](../CONTRIBUTING.md) - How to contribute (coming soon)

Questions? Open an issue on GitHub!
