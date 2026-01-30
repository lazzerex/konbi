# Konbi

<p align="center">
  

<img width="811" height="308" alt="konbi_fix_logo" src="https://github.com/user-attachments/assets/91e4068e-4738-400b-9984-eac70c189cb8" />


</p>

<p align="center">
  <img src="https://img.shields.io/badge/Go-1.21-00ADD8?style=flat&logo=go&logoColor=white"/>
  <img src="https://img.shields.io/badge/React-18.2-61DAFB?style=flat&logo=react&logoColor=black"/>
  <img src="https://img.shields.io/badge/Database-Neon-00E699?style=flat&logo=postgresql&logoColor=white"/>
  <img src="https://img.shields.io/badge/Backend-Railway-7B3FE4?style=flat&logo=railway&logoColor=white"/>
  <img src="https://img.shields.io/badge/Deployed%20on-Vercel-000000?style=flat&logo=vercel&logoColor=white"/>
  <img src="https://img.shields.io/badge/License-MIT-F7DF1E?style=flat&logoColor=black"/>
</p>



<p align="center">
  <strong>A minimal, elegant web application for sharing files and text notes online with zero friction. Upload files or create notes and instantly get a shareable link.</strong>
</p>

<p align="center">
  <strong>Live Application</strong><br/>
  Frontend deployed on <strong>Vercel</strong> | Backend deployed on <strong>Railway</strong> | Database managed with <strong>Neon</strong>
</p>
<img width="1184" height="693" alt="konbi1" src="https://github.com/user-attachments/assets/3c52e1a3-f29a-4e0a-b790-99dca741f4ef" />
<img width="1177" height="689" alt="konbi3" src="https://github.com/user-attachments/assets/79eaef47-a80d-4ff9-98a9-3da43f41b3e1" />
<img width="1171" height="684" alt="konbi2" src="https://github.com/user-attachments/assets/55840d19-9d99-4f5a-a566-b462558796f3" />
<img width="1169" height="689" alt="konbi4" src="https://github.com/user-attachments/assets/a09129d4-fed7-4ed9-a5f7-c749fb5a4d32" />



## Features

**Easy Sharing**
- Drag and drop file uploads
- Simple text note creation
- Instant shareable links with unique IDs
- One-click copy to clipboard

**Security & Limits**
- Files up to 50MB
- Content expires after 7 days (configurable)
- Rate limiting (10 req/second)
- File type validation

**Analytics**
- View count tracking
- Upload statistics
- Expiration timestamps

## Tech Stack

**Frontend:**
- React 18
- Framer Motion for animations
- Axios for API calls
- Responsive design with theme support
- Deployed on Vercel

**Backend:**
- Go (Golang)
- Gin web framework
- PostgreSQL (Neon) or SQLite database
- Local filesystem storage
- Deployed on Railway

## Project Structure

```
konbi/
├── backend/
│   ├── main.go                          # Application entry point
│   ├── internal/
│   │   ├── config/                      # Configuration management
│   │   │   └── config.go
│   │   ├── models/                      # Data structures
│   │   │   └── content.go
│   │   ├── errors/                      # Custom error types
│   │   │   └── errors.go
│   │   ├── repository/                  # Database layer
│   │   │   ├── db.go
│   │   │   └── content_repository.go
│   │   ├── services/                    # Business logic
│   │   │   └── content_service.go
│   │   ├── handlers/                    # HTTP handlers
│   │   │   ├── content_handler.go
│   │   │   └── health.go
│   │   └── middleware/                  # Custom middleware
│   │       ├── auth.go
│   │       ├── logger.go
│   │       └── ratelimit.go
│   ├── uploads/                         # Uploaded files storage
│   ├── go.mod                           # Go dependencies
│   ├── Dockerfile                       # Backend Docker config
│   └── railway.toml                     # Railway deployment config
├── frontend/
│   ├── public/
│   │   └── index.html
│   ├── src/
│   │   ├── components/
│   │   │   ├── ShareMode.js     # Upload/create interface
│   │   │   ├── ShareMode.css
│   │   │   ├── AccessMode.js    # Retrieve interface
│   │   │   └── AccessMode.css
│   │   ├── App.js
│   │   ├── App.css
│   │   ├── config.js            # API URL config
│   │   └── index.js
│   ├── package.json
│   ├── Dockerfile
│   └── vercel.json              # Vercel deployment config
└── docker-compose.yml           # Local development setup
```

## Architecture

The backend follows **clean architecture** principles with clear separation of concerns:

```
┌─────────────┐
│  HTTP Layer │  ← Handlers (routes, request/response)
└──────┬──────┘
       │
┌──────▼──────┐
│ Service Layer│  ← Business logic, validation
└──────┬──────┘
       │
┌──────▼──────┐
│Repository   │  ← Database operations
└──────┬──────┘
       │
┌──────▼──────┐
│  Database   │  ← PostgreSQL/SQLite
└─────────────┘
```

**Key Features:**
- **Structured Logging**: All requests logged with request ID, latency, status
- **Custom Error Types**: Proper HTTP status codes and error messages
- **Middleware**: Rate limiting, logging, authentication
- **Database Abstraction**: Works with both PostgreSQL and SQLite
- **Connection Pooling**: Optimized database connections
- **Graceful Shutdown**: Proper cleanup on server stop

## Getting Started

### Using the Live Application

The application is live and ready to use:
- Visit the deployed frontend on Vercel
- Share files up to 50MB or create text notes
- Get instant shareable links
- Content automatically expires after 7 days

### Local Development Setup

If you want to run the application locally or contribute to development:

### Prerequisites

- Go 1.21 or higher
- Node.js 18 or higher
- npm or yarn

### Backend Setup

1. Navigate to the backend directory:
```bash
cd backend
```

2. Install Go dependencies:
```bash
go mod download
```

3. Run the backend server:
```bash
# Simple run
go run .

# With environment variables
ENVIRONMENT=development PORT=8080 ADMIN_SECRET=your-secret go run .
```

The server will start on `http://localhost:8080`

**Environment Variables:**
- `DATABASE_URL` - PostgreSQL connection string (for Neon/PostgreSQL)
- `PORT` - Server port (default: 8080)
- `ENVIRONMENT` - Environment mode: development or production (default: development)
- `ADMIN_SECRET` - Secret for admin endpoints (optional)
- `ALLOWED_ORIGINS` - CORS allowed origins (default: http://localhost:3000)
- `MAX_FILE_SIZE_MB` - Max upload size in MB (default: 50)
- `EXPIRATION_DAYS` - Content expiration time (default: 7)
- `RATE_LIMIT_PER_SEC` - Rate limit per second (default: 10)
- `DB_MAX_CONNECTIONS` - Max database connections (default: 25)
- `DB_MAX_IDLE_CONNS` - Max idle connections (default: 5)

### Frontend Setup

1. Navigate to the frontend directory:
```bash
cd frontend
```

2. Install dependencies:
```bash
npm install
```

3. Create a `.env` file:
```bash
cp .env.example .env
```

4. Update `.env` with your backend URL:
```
REACT_APP_API_URL=http://localhost:8080/api
```

5. Start the development server:
```bash
npm start
```

The app will open at `http://localhost:3000`

### Using Docker Compose

Run both frontend and backend together:

```bash
docker-compose up --build
```

Access:
- Frontend: http://localhost:3000
- Backend: http://localhost:8080

## API Documentation

### POST `/api/upload`
Upload a file and get a share ID.

**Request:**
- Content-Type: `multipart/form-data`
- Body: `file` (file data)

**Response:**
```json
{
  "id": "AbC123Xy",
  "filename": "document.pdf",
  "size": 1048576,
  "expiresAt": "2026-02-02T12:00:00Z"
}
```

### POST `/api/note`
Create a text note and get a share ID.

**Request:**
```json
{
  "title": "My Note",
  "content": "Note content here..."
}
```

**Response:**
```json
{
  "id": "DeF456Zw",
  "title": "My Note",
  "expiresAt": "2026-02-02T12:00:00Z"
}
```

### GET `/api/content/:id`
Retrieve content by ID.

**Response (File):**
```json
{
  "type": "file",
  "filename": "document.pdf",
  "size": 1048576,
  "downloadUrl": "/api/content/AbC123Xy/download"
}
```

**Response (Note):**
```json
{
  "type": "note",
  "title": "My Note",
  "content": "Note content here..."
}
```

### GET `/api/stats/:id`
Get statistics for content.

**Response:**
```json
{
  "viewCount": 42,
  "createdAt": "2026-01-26T12:00:00Z",
  "expiresAt": "2026-02-02T12:00:00Z"
}
```

## Deployment

The application is currently deployed and running in production:
- **Frontend:** Vercel (automatic deployments from main branch)
- **Backend:** Railway (automatic deployments from main branch)

### Deploying Your Own Instance

If you want to deploy your own instance of Konbi, follow these guides:

### Frontend (Vercel)

1. Install Vercel CLI (optional, can also deploy via web):
```bash
npm install -g vercel
```

2. Connect your GitHub repository to Vercel:
   - Go to [Vercel Dashboard](https://vercel.com)
   - Click "New Project"
   - Import your repository
   - Select the `frontend` directory as the root

3. Configure build settings:
   - **Framework Preset:** Create React App
   - **Build Command:** `npm run build`
   - **Output Directory:** `build`

4. Set environment variable in Vercel dashboard:
```
REACT_APP_API_URL=https://your-railway-backend-url.up.railway.app/api
```

5. Deploy - Vercel will automatically deploy on every push to main branch

### Backend (Railway)

**Recommended for Production**

1. Push your code to GitHub

2. Create a free Neon database:
   - Go to [neon.tech](https://neon.tech) and create an account
   - Create a new project
   - Copy the connection string

3. Go to [Railway.app](https://railway.app)

4. Click "New Project" → "Deploy from GitHub repo"

5. Select your repository

6. Railway will automatically:
   - Detect the Go application
   - Use the `railway.toml` configuration
   - Set up the build and deployment

7. Configure environment variables in Railway dashboard:
   - `DATABASE_URL` - Your Neon PostgreSQL connection string
   - `ADMIN_SECRET` - Random secret for admin endpoints (optional)
   - `ALLOWED_ORIGINS` - Your Vercel frontend URL

8. Your backend will be live with a Railway URL

**Note:** Using Neon database provides better performance and data persistence compared to SQLite. The backend automatically detects and uses PostgreSQL when `DATABASE_URL` is set.

### Backend (Fly.io)

**Alternative deployment option:**

1. Install Fly CLI:
```bash
curl -L https://fly.io/install.sh | sh
```

2. Navigate to backend directory:
```bash
cd backend
```

3. Launch app:
```bash
fly launch
```

4. Deploy:
```bash
fly deploy
```

5. Create persistent volume:
```bash
fly volumes create konbi_data --size 1
```

### Backend (Render)

**Alternative deployment option:**

1. Create new Web Service on Render
2. Connect your GitHub repository
3. Configure:
   - **Build Command:** `go build -o konbi`
   - **Start Command:** `./konbi`
   - **Environment:** Go
4. Add environment variables if needed

## Production Architecture

The live application uses the following architecture:

```
User Browser
     ↓
Vercel (Frontend - React SPA)
     ↓
Railway (Backend - Go API)
     ↓
Neon (PostgreSQL Database)
     ↓
File Storage (Railway Volume/Local)
```

**Key Features in Production:**
- HTTPS encryption on both frontend and backend
- CORS configured for secure cross-origin requests
- Rate limiting to prevent abuse (10 req/sec)
- Automatic HTTPS redirects
- CDN distribution via Vercel
- Serverless PostgreSQL database via Neon
- Auto-scaling database with Neon
- Protected admin endpoints with secret authentication
- Automatic deployments from GitHub
- Security: npm packages updated and vulnerabilities patched

## Configuration

### Expiration Time
Change content expiration in [backend/internal/config/config.go](backend/internal/config/config.go):
```go
ExpirationDays: getEnvAsInt("EXPIRATION_DAYS", 7),  // Change default value
```

Or set via environment variable:
```bash
EXPIRATION_DAYS=14 go run .
```

### File Size Limit
Modify in [backend/internal/config/config.go](backend/internal/config/config.go):
```go
MaxFileSize: int64(getEnvAsInt("MAX_FILE_SIZE_MB", 50)) * 1024 * 1024,
```

Or set via environment variable:
```bash
MAX_FILE_SIZE_MB=100 go run .
```

### Allowed File Types
Edit the `allowedExtensions` map in [backend/internal/services/content_service.go](backend/internal/services/content_service.go):
```go
var allowedExtensions = map[string]bool{
    ".txt": true,
    ".pdf": true,
    // Add more extensions...
}
```

### Rate Limiting
Adjust in [backend/internal/config/config.go](backend/internal/config/config.go) or set via environment:
```bash
RATE_LIMIT_PER_SEC=20 RATE_LIMIT_BURST=20 go run .
```

## Database

The app supports both **PostgreSQL** (production) and **SQLite** (development):

- **PostgreSQL (Neon)**: Automatically used when `DATABASE_URL` environment variable is set
- **SQLite**: Used for local development when `DATABASE_URL` is not set

The database schema includes:

```sql
CREATE TABLE content (
    id TEXT PRIMARY KEY,
    type TEXT NOT NULL,           -- 'file' or 'note'
    title TEXT,
    filename TEXT,
    filepath TEXT,
    filesize INTEGER,             -- BIGINT in PostgreSQL
    content TEXT,
    created_at DATETIME,          -- TIMESTAMP in PostgreSQL
    expires_at DATETIME NOT NULL, -- TIMESTAMP in PostgreSQL
    view_count INTEGER DEFAULT 0,
    deleted_at DATETIME           -- TIMESTAMP in PostgreSQL, for soft deletes
);

-- Performance indexes
CREATE INDEX idx_content_expires_at ON content(expires_at);
CREATE INDEX idx_content_deleted_at ON content(deleted_at);
CREATE INDEX idx_content_type ON content(type);
CREATE INDEX idx_content_created_at ON content(created_at DESC);
```

**Automatic Migrations**: The application automatically creates tables and indexes on startup. Database migrations run in [backend/internal/repository/db.go](backend/internal/repository/db.go).

**Soft Deletes**: Content is marked as deleted (not permanently removed) for data retention and audit trails.

### Using Neon PostgreSQL (Recommended)

1. Create free Neon account at [neon.tech](https://neon.tech)
2. Create a new project (Postgres 17 recommended)
3. Copy the connection string
4. Set as environment variable:
   ```bash
   export DATABASE_URL="postgresql://user:pass@host.neon.tech/dbname?sslmode=require"
   ```
5. Run your backend - it will automatically use PostgreSQL

**Benefits of Neon:**
- Free tier with 0.5GB storage
- Auto-scaling and serverless
- Better performance for concurrent requests
- Automatic backups
- Database branching for development
- No vendor lock-in (migrate to any hosting platform)

### Using SQLite (Development)

For local development, simply run without `DATABASE_URL`:
```bash
cd backend
go run .
```

The backend will automatically create and use `konbi.db` locally.

## Maintenance

### Manual Cleanup

Delete expired content manually:

**SQLite (local):**
```bash
sqlite3 konbi.db "DELETE FROM content WHERE expires_at < datetime('now')"
```

**PostgreSQL (via psql):**
```bash
psql $DATABASE_URL -c "DELETE FROM content WHERE expires_at < now()"
```

**Via Admin API:**
```bash
# First, view all content
curl -H "X-Admin-Secret: your-secret" \
  https://your-backend.up.railway.app/api/admin/list
```

### Backup Database

**SQLite:**
```bash
sqlite3 konbi.db ".backup konbi_backup.db"
```

**PostgreSQL (Neon):**
Neon provides automatic backups with point-in-time recovery. You can also export manually:
```bash
pg_dump $DATABASE_URL > konbi_backup.sql
```

### View All Content

**SQLite:**
```bash
sqlite3 konbi.db "SELECT id, type, filename, title, created_at, expires_at FROM content"
```

**PostgreSQL:**
```bash
psql $DATABASE_URL -c "SELECT id, type, filename, title, created_at, expires_at FROM content"
```

## Development

### Running Tests

Backend:
```bash
cd backend
go test ./...
```

Frontend:
```bash
cd frontend
npm test
```

### Building for Production

Backend:
```bash
cd backend
CGO_ENABLED=1 go build -o konbi
```

Frontend:
```bash
cd frontend
npm run build
```

## Troubleshooting

### CORS Issues

If you encounter CORS errors, update the allowed origins in [backend/main.go](backend/main.go):
```go
config.AllowOrigins = []string{"https://your-frontend-domain.com"}
```

### Database Locked

If SQLite database is locked:
```bash
rm konbi.db-shm konbi.db-wal
```

### Port Already in Use

Change the port:
```bash
PORT=3001 go run .  # Backend
PORT=3002 npm start  # Frontend
```

### npm Vulnerabilities

The frontend dependencies are kept up-to-date with security patches. If you see warnings:
```bash
cd frontend
npm audit
```

The project uses npm overrides to patch vulnerable dependencies in react-scripts.

## Migration Guide

### From Railway to Fly.io/Render

Since the database is on Neon (separate from hosting), migration is simple:

1. Deploy backend to new platform (Fly.io/Render)
2. Set the **same** `DATABASE_URL` environment variable
3. Update frontend's `REACT_APP_API_URL` to new backend URL
4. Done! No data migration needed.

### From SQLite to Neon

If you're currently using SQLite and want to migrate to Neon:

1. Export SQLite data:
   ```bash
   sqlite3 konbi.db .dump > export.sql
   ```

2. Create Neon database and set `DATABASE_URL`

3. The backend will automatically create tables on startup

4. Import data if needed (manual process for schema differences)

Note: Since content expires after 7 days, starting fresh is often acceptable.

## Security

- **Admin Endpoints**: Protected by `X-Admin-Secret` header authentication
- **Rate Limiting**: 10 requests per second per IP
- **File Validation**: Whitelisted file extensions only
- **Size Limits**: 50MB max file size
- **CORS**: Configured for specific origins in production
- **Dependencies**: Regularly updated, security vulnerabilities patched
- **HTTPS**: Enforced in production via Vercel and Railway

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

MIT License - feel free to use this project for any purpose.

## Support

For issues and questions, please open an issue on GitHub.

---

Built with Go and React
