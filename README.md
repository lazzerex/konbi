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

<p align="center">
  <strong>Live Application</strong><br/>
  Frontend deployed on <strong>Vercel</strong> | Backend deployed on <strong>Railway</strong>
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
- SQLite database
- Local filesystem storage
- Deployed on Railway

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

2. Go to [Railway.app](https://railway.app)

3. Click "New Project" → "Deploy from GitHub repo"

4. Select your repository

5. Railway will automatically:
   - Detect the Go application
   - Use the `railway.toml` configuration
   - Set up the build and deployment

6. Configure environment variables (optional):
   - `PORT` - Railway sets this automatically
   - `DB_PATH` - Path for SQLite database

7. Your backend will be live with a Railway URL

**Note:** Railway provides persistent volumes for SQLite database storage.

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
SQLite Database (Persistent Volume)
     ↓
File Storage (Railway Volume)
```

**Key Features in Production:**
- HTTPS encryption on both frontend and backend
- CORS configured for secure cross-origin requests
- Rate limiting to prevent abuse
- Automatic HTTPS redirects
- CDN distribution via Vercel
- Persistent storage on Railway
- Automatic deployments from GitHub

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
