package main

import (
	"database/sql"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/time/rate"
)

var (
	db      *sql.DB
	limiter = rate.NewLimiter(rate.Every(time.Second), 10) // 10 requests per second
)

func main() {
	var err error

	// init database
	db, err = initDB()
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// ensure uploads directory exists
	if err := os.MkdirAll("uploads", 0755); err != nil {
		log.Fatalf("Failed to create uploads directory: %v", err)
	}

	// start cleanup routine
	go cleanupExpiredContent()

	// setup router
	r := gin.Default()

	// cors config
	config := cors.DefaultConfig()
	config.AllowOrigins = []string{"*"} // configure for production
	config.AllowMethods = []string{"GET", "POST", "OPTIONS"}
	config.AllowHeaders = []string{"Origin", "Content-Type", "Content-Length", "Accept"}
	r.Use(cors.New(config))

	// rate limiting
	r.Use(rateLimitMiddleware())

	// api routes
	api := r.Group("/api")
	{
		api.POST("/upload", handleUpload)
		api.POST("/note", handleNote)
		api.GET("/content/:id", handleGetContent)
		api.GET("/content/:id/download", handleDownload)
		api.GET("/stats/:id", handleGetStats)
	}

	// health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on port %s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func initDB() (*sql.DB, error) {
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "./konbi.db"
	}

	database, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}

	// create tables
	schema := `
	CREATE TABLE IF NOT EXISTS content (
		id TEXT PRIMARY KEY,
		type TEXT NOT NULL,
		title TEXT,
		filename TEXT,
		filepath TEXT,
		filesize INTEGER,
		content TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		expires_at DATETIME NOT NULL,
		view_count INTEGER DEFAULT 0
	);

	CREATE INDEX IF NOT EXISTS idx_expires_at ON content(expires_at);
	`

	if _, err := database.Exec(schema); err != nil {
		return nil, err
	}

	return database, nil
}

func rateLimitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !limiter.Allow() {
			c.JSON(429, gin.H{"error": "Rate limit exceeded"})
			c.Abort()
			return
		}
		c.Next()
	}
}

func cleanupExpiredContent() {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for range ticker.C {
		log.Println("Running cleanup routine...")

		rows, err := db.Query(`
			SELECT id, filepath FROM content 
			WHERE expires_at < ? AND type = 'file'
		`, time.Now().UTC().Format("2006-01-02 15:04:05"))
		if err != nil {
			log.Printf("Cleanup query error: %v", err)
			continue
		}

		var deleted int
		for rows.Next() {
			var id, filepath string
			if err := rows.Scan(&id, &filepath); err != nil {
				log.Printf("Scan error: %v", err)
				continue
			}

			// delete file
			if err := os.Remove(filepath); err != nil && !os.IsNotExist(err) {
				log.Printf("Failed to delete file %s: %v", filepath, err)
			}

			deleted++
		}
		rows.Close()

		// delete expired records
		result, err := db.Exec(`DELETE FROM content WHERE expires_at < ?`, time.Now().UTC().Format("2006-01-02 15:04:05"))
		if err != nil {
			log.Printf("Failed to delete expired content: %v", err)
		} else {
			count, _ := result.RowsAffected()
			log.Printf("Cleaned up %d expired items (%d files deleted)", count, deleted)
		}
	}
}

func getFileExtension(filename string) string {
	ext := filepath.Ext(filename)
	if ext == "" {
		return ".bin"
	}
	return ext
}
