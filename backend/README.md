# Backend Setup

## Quick Start

```bash
# Install dependencies
go mod download

# Run the server
go run .
```

The server will start on http://localhost:8080

## Environment Variables

Create a `.env` file or set these in your environment:

```bash
PORT=8080                    # Server port
DB_PATH=./konbi.db          # SQLite database path
```

## API Endpoints

### Upload File
```bash
curl -X POST http://localhost:8080/api/upload \
  -F "file=@/path/to/file.pdf"
```

### Create Note
```bash
curl -X POST http://localhost:8080/api/note \
  -H "Content-Type: application/json" \
  -d '{"title":"My Note","content":"Hello World"}'
```

### Get Content
```bash
curl http://localhost:8080/api/content/AbC123Xy
```

### Get Stats
```bash
curl http://localhost:8080/api/stats/AbC123Xy
```

## Build for Production

```bash
CGO_ENABLED=1 go build -o konbi
./konbi
```

## Database Management

### View all content
```bash
sqlite3 konbi.db "SELECT * FROM content"
```

### Manual cleanup
```bash
sqlite3 konbi.db "DELETE FROM content WHERE expires_at < datetime('now')"
```

### Backup
```bash
sqlite3 konbi.db ".backup konbi_backup.db"
```

## Configuration

Edit constants in `handlers.go`:
- `maxFileSize` - Max upload size (default 50MB)
- `expirationDays` - Content expiration time (default 7 days)
- `allowedExtensions` - Allowed file types

Edit rate limiting in `main.go`:
- `rate.NewLimiter(rate.Every(time.Second), 10)` - 10 requests per second

## Deployment

### Railway
```bash
railway login
railway init
railway up
```

### Fly.io
```bash
fly launch
fly deploy
fly volumes create konbi_data --size 1
```

### Docker
```bash
docker build -t konbi-backend .
docker run -p 8080:8080 -v $(pwd)/uploads:/root/uploads konbi-backend
```

## Monitoring

### Health check
```bash
curl http://localhost:8080/health
```

### View logs
```bash
# Development
go run . 2>&1 | tee server.log

# Production
./konbi 2>&1 | tee server.log
```
