# Getting Started - No External Dependencies Required!

This guide will help you get the News Aggregator running without installing PostgreSQL, Redis, RabbitMQ, or Elasticsearch manually.

## 🐳 Using Docker (Recommended)

Docker will handle all the infrastructure for you automatically.

### Step 1: Install Docker Desktop

Download and install Docker Desktop for your operating system:
- **macOS**: https://docs.docker.com/desktop/install/mac-install/
- **Windows**: https://docs.docker.com/desktop/install/windows-install/
- **Linux**: https://docs.docker.com/desktop/install/linux-install/

### Step 2: Verify Docker Installation

```bash
# Check if Docker is installed and running
docker --version
docker-compose --version

# You should see version numbers like:
# Docker version 24.0.x
# Docker Compose version v2.x.x
```

### Step 3: Start the News Aggregator

```bash
# Navigate to the project directory
cd /Users/pkumar495/Documents/NEWSAggregator

# Start all services (first time will download images - may take 5-10 minutes)
docker-compose up -d

# Check if services are starting
docker-compose ps
```

### Step 4: Wait for Services to Initialize

The first startup takes a few minutes because:
- Docker downloads the required images
- PostgreSQL initializes the database
- Elasticsearch sets up indices
- Go applications compile and start

```bash
# Watch the logs to see progress
docker-compose logs -f

# Or check specific service logs
docker-compose logs -f api-gateway
docker-compose logs -f postgres
```

### Step 5: Verify Everything is Working

```bash
# Check health endpoint (wait until this returns healthy)
curl http://localhost:8080/health

# Test the API
curl http://localhost:8080/api/v1/news
curl http://localhost:8080/api/v1/categories

# Register a test user
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "username": "testuser",
    "password": "password123",
    "first_name": "Test",
    "last_name": "User"
  }'
```

### Step 6: Access the Services

Once everything is running, you can access:

- **API**: http://localhost:8080
- **Health Check**: http://localhost:8080/health
- **Grafana Dashboard**: http://localhost:3000 (admin/admin)
- **Prometheus Metrics**: http://localhost:9090
- **RabbitMQ Management**: http://localhost:15672 (guest/guest)

## 🛠️ Development Mode

If you want to develop and make changes to the code:

### Option A: Hybrid (Infrastructure in Docker, Apps Local)

```bash
# Start only infrastructure services
docker-compose up -d postgres redis rabbitmq elasticsearch prometheus grafana

# Run Go applications locally for development
go run ./cmd/api-gateway
# In another terminal:
go run ./cmd/data-collector
# In another terminal:
go run ./cmd/processor
```

### Option B: Full Docker with Live Reload

You can modify the Dockerfiles to use air for live reloading during development.

## 🔧 Useful Commands

```bash
# Start all services
docker-compose up -d

# Stop all services
docker-compose down

# View logs
docker-compose logs -f

# Restart a specific service
docker-compose restart api-gateway

# Rebuild and restart after code changes
docker-compose up -d --build

# Clean up everything (removes data!)
docker-compose down -v --remove-orphans

# Check service status
docker-compose ps

# Execute commands in containers
docker-compose exec postgres psql -U postgres -d news_aggregator
docker-compose exec redis redis-cli
```

## 🐛 Troubleshooting

### Services Won't Start
```bash
# Check if ports are already in use
lsof -i :8080  # API Gateway
lsof -i :5432  # PostgreSQL
lsof -i :6379  # Redis
lsof -i :5672  # RabbitMQ

# Kill processes using those ports if needed
kill -9 <PID>
```

### Database Connection Issues
```bash
# Check PostgreSQL logs
docker-compose logs postgres

# Connect to database manually
docker-compose exec postgres psql -U postgres -d news_aggregator
```

### Out of Memory/Disk Space
```bash
# Clean up Docker system
docker system prune -a

# Check disk usage
docker system df
```

### Service Health Check Fails
```bash
# Check individual service health
curl http://localhost:8080/health
docker-compose exec postgres pg_isready -U postgres
docker-compose exec redis redis-cli ping
```

## 📊 Monitoring Your Application

### View Application Metrics
1. Open Grafana: http://localhost:3000
2. Login: admin/admin
3. Import dashboards or create custom ones

### View Message Queue
1. Open RabbitMQ Management: http://localhost:15672
2. Login: guest/guest
3. Monitor queues and message flow

### Check Database
```bash
# Connect to PostgreSQL
docker-compose exec postgres psql -U postgres -d news_aggregator

# List tables
\dt

# Check news articles
SELECT COUNT(*) FROM news;
SELECT title, source, category FROM news LIMIT 5;
```

## 🚀 What's Next?

1. **Test the API**: Use the curl commands above or a tool like Postman
2. **Add News Sources**: Modify `configs/config.yaml` to add your preferred news sources
3. **Customize**: Modify the code and rebuild with `docker-compose up -d --build`
4. **Scale**: Add more instances of services in the docker-compose.yml
5. **Deploy**: Use the same Docker setup for production deployment

## 💡 Tips

- **First Run**: The first `docker-compose up` takes longer due to image downloads
- **Data Persistence**: Your data is stored in Docker volumes and persists between restarts
- **Development**: Use `docker-compose logs -f api-gateway` to watch your application logs
- **Configuration**: Environment variables in docker-compose.yml override config.yaml settings
- **Cleanup**: Use `docker-compose down -v` to completely reset (this deletes all data!)

That's it! You now have a fully functional news aggregator running without installing any external dependencies manually.
