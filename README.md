# Konbi

<p align="center">
  

<img width="811" height="308" alt="konbi_fix_logo" src="https://github.com/user-attachments/assets/91e4068e-4738-400b-9984-eac70c189cb8" />


</p>

<p align="center">
  <img src="https://img.shields.io/badge/Go-1.21-00ADD8?style=flat&logo=go&logoColor=white"/>
  <img src="https://img.shields.io/badge/React-18.2-61DAFB?style=flat&logo=react&logoColor=black"/>
  <img src="https://img.shields.io/badge/License-MIT-green?style=flat"/>
  <img src="https://img.shields.io/badge/Deployed%20on-Vercel-black?style=flat&logo=vercel&logoColor=white"/>
  <img src="https://img.shields.io/badge/Backend-Railway-0B0D0E?style=flat&logo=railway&logoColor=white"/>
</p>

<p align="center">
  <strong>A minimal, elegant web application for sharing files and text notes online with zero friction. Upload files or create notes and instantly get a shareable link.</strong>
</p>

> **Note:** Current version runs on local machine. Production deployment coming soon.

<img width="892" height="658" alt="Screenshot_2026-01-26_23-30-00" src="https://github.com/user-attachments/assets/c0bd1543-61ba-4d66-ac13-5d187c0d06d3" />

<img width="889" height="634" alt="Screenshot_2026-01-26_23-30-34" src="https://github.com/user-attachments/assets/d5ac8449-057f-4854-9b2a-f15b7e209923" />

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
- Axios for API calls
- Responsive CSS design
- Hosted on Vercel

**Backend:**
- Go (Golang)
- Gin web framework
- SQLite database
- Local filesystem storage

## Project Structure

```
konbi/
├── backend/
│   ├── main.go              # Server setup and initialization
│   ├── handlers.go          # API endpoint handlers
│   ├── go.mod               # Go dependencies
│   ├── Dockerfile           # Backend Docker config
│   └── railway.toml         # Railway deployment config
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

## Getting Started

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
go run .
```

The server will start on `http://localhost:8080`

**Environment Variables:**
- `PORT` - Server port (default: 8080)
- `DB_PATH` - Database file path (default: ./konbi.db)

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

### Frontend (Vercel)

1. Install Vercel CLI:
```bash
npm install -g vercel
```

2. Navigate to frontend directory:
```bash
cd frontend
```

3. Deploy:
```bash
vercel --prod
```

4. Set environment variable in Vercel dashboard:
```
REACT_APP_API_URL=https://your-backend-url.com/api
```

### Backend (Railway)

1. Install Railway CLI:
```bash
npm install -g @railway/cli
```

2. Login and initialize:
```bash
railway login
cd backend
railway init
```

3. Deploy:
```bash
railway up
```

4. The service will automatically use `railway.toml` configuration

### Backend (Fly.io)

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

1. Create new Web Service on Render
2. Connect your GitHub repository
3. Configure:
   - **Build Command:** `go build -o konbi`
   - **Start Command:** `./konbi`
   - **Environment:** Go
4. Add environment variables if needed

## Configuration

### Expiration Time
Change content expiration in [backend/handlers.go](backend/handlers.go):
```go
const expirationDays = 7  // Change to desired number of days
```

### File Size Limit
Modify in [backend/handlers.go](backend/handlers.go):
```go
const maxFileSize = 50 * 1024 * 1024  // 50MB in bytes
```

### Allowed File Types
Edit the `allowedExtensions` map in [backend/handlers.go](backend/handlers.go):
```go
var allowedExtensions = map[string]bool{
    ".txt": true,
    ".pdf": true,
    // Add more extensions...
}
```

### Rate Limiting
Adjust in [backend/main.go](backend/main.go):
```go
limiter = rate.NewLimiter(rate.Every(time.Second), 10)  // 10 requests per second
```

## Database

The app uses SQLite by default. The database schema includes:

```sql
CREATE TABLE content (
    id TEXT PRIMARY KEY,
    type TEXT NOT NULL,           -- 'file' or 'note'
    title TEXT,
    filename TEXT,
    filepath TEXT,
    filesize INTEGER,
    content TEXT,
    created_at DATETIME,
    expires_at DATETIME NOT NULL,
    view_count INTEGER DEFAULT 0
);
```

### Migrating to PostgreSQL

To use PostgreSQL instead of SQLite:

1. Install PostgreSQL driver:
```bash
go get github.com/lib/pq
```

2. Update imports in `main.go`:
```go
import _ "github.com/lib/pq"
```

3. Change connection string:
```go
db, err := sql.Open("postgres", "postgres://user:pass@host:5432/dbname?sslmode=disable")
```

## Maintenance

### Manual Cleanup

Delete expired content manually:
```bash
# In backend directory
sqlite3 konbi.db "DELETE FROM content WHERE expires_at < datetime('now')"
```

### Backup Database

```bash
sqlite3 konbi.db ".backup konbi_backup.db"
```

### View All Content

```bash
sqlite3 konbi.db "SELECT id, type, filename, title, created_at, expires_at FROM content"
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

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

MIT License - feel free to use this project for any purpose.

## Support

For issues and questions, please open an issue on GitHub.

---

Built with Go and React
