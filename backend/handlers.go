package main

import (
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// Helper to convert SQLite ? placeholders to PostgreSQL $1, $2, etc.
func toSQL(query string) string {
	// Check if using PostgreSQL
	if os.Getenv("DATABASE_URL") != "" {
		// Convert ? to $1, $2, $3...
		result := ""
		paramCount := 1
		for _, char := range query {
			if char == '?' {
				result += fmt.Sprintf("$%d", paramCount)
				paramCount++
			} else {
				result += string(char)
			}
		}
		// Replace SQLite datetime('now') with PostgreSQL now()
		result = strings.ReplaceAll(result, "datetime('now')", "now()")
		return result
	}
	return query
}

const (
	maxFileSize    = 50 * 1024 * 1024 // 50MB
	expirationDays = 7
	idLength       = 8
)

var allowedExtensions = map[string]bool{
	".txt": true, ".pdf": true, ".doc": true, ".docx": true,
	".jpg": true, ".jpeg": true, ".png": true, ".gif": true,
	".zip": true, ".tar": true, ".gz": true,
	".mp4": true, ".mp3": true, ".wav": true,
	".csv": true, ".xlsx": true, ".xls": true,
	".json": true, ".xml": true, ".md": true,
}

func generateID() (string, error) {
	bytes := make([]byte, 8)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}

	// base64 encode and make url-safe
	id := base64.RawURLEncoding.EncodeToString(bytes)
	// use only alphanumeric
	id = strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') {
			return r
		}
		return -1
	}, id)

	// if too short, try again
	if len(id) < idLength {
		return generateID()
	}

	return id[:idLength], nil
}

func handleUpload(c *gin.Context) {
	// parse multipart form
	if err := c.Request.ParseMultipartForm(maxFileSize); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File too large or invalid request"})
		return
	}

	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No file provided"})
		return
	}
	defer file.Close()

	// validate file size
	if header.Size > maxFileSize {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("File size exceeds %dMB limit", maxFileSize/1024/1024)})
		return
	}

	// validate file extension
	ext := strings.ToLower(filepath.Ext(header.Filename))
	if ext != "" && !allowedExtensions[ext] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File type not allowed"})
		return
	}

	// Generate unique ID
	id, err := generateID()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate ID"})
		return
	}

	// Check for ID collision
	var exists bool
	err = db.QueryRow(toSQL("SELECT EXISTS(SELECT 1 FROM content WHERE id = ?)"), id).Scan(&exists)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}
	if exists {
		// retry with new id
		handleUpload(c)
		return
	}

	// save file
	filename := fmt.Sprintf("%s%s", id, ext)
	filePath := filepath.Join("uploads", filename)

	dst, err := os.Create(filePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file"})
		return
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		os.Remove(filePath)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file"})
		return
	}

	// save metadata to db
	expiresAt := time.Now().UTC().Add(expirationDays * 24 * time.Hour)
	_, err = db.Exec(toSQL(`
		INSERT INTO content (id, type, filename, filepath, filesize, expires_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`), id, "file", header.Filename, filePath, header.Size, expiresAt.Format("2006-01-02 15:04:05"))

	if err != nil {
		os.Remove(filePath)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save metadata"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":        id,
		"filename":  header.Filename,
		"size":      header.Size,
		"expiresAt": expiresAt.Format(time.RFC3339),
	})
}

func handleNote(c *gin.Context) {
	var req struct {
		Title   string `json:"title"`
		Content string `json:"content" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	// validate content length
	if len(req.Content) > 1024*1024 { // 1mb limit
		c.JSON(http.StatusBadRequest, gin.H{"error": "Content too large"})
		return
	}

	// generate unique id
	id, err := generateID()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate ID"})
		return
	}

	// check for id collision
	var exists bool
	err = db.QueryRow(toSQL("SELECT EXISTS(SELECT 1 FROM content WHERE id = ?)"), id).Scan(&exists)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}
	if exists {
		// retry with new id
		handleNote(c)
		return
	}

	// save to db
	expiresAt := time.Now().UTC().Add(expirationDays * 24 * time.Hour)
	_, err = db.Exec(toSQL(`
		INSERT INTO content (id, type, title, content, expires_at)
		VALUES (?, ?, ?, ?, ?)
	`), id, "note", req.Title, req.Content, expiresAt.Format("2006-01-02 15:04:05"))

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save note"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":        id,
		"title":     req.Title,
		"expiresAt": expiresAt.Format(time.RFC3339),
	})
}

func handleGetContent(c *gin.Context) {
	id := c.Param("id")

	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID required"})
		return
	}

	var contentType string
	var title, filename, filePath, content sql.NullString
	var filesize sql.NullInt64

	err := db.QueryRow(toSQL(`
		SELECT type, title, filename, filepath, filesize, content
		FROM content
		WHERE id = ? AND expires_at > ?
	`), id, time.Now().UTC().Format("2006-01-02 15:04:05")).Scan(&contentType, &title, &filename, &filePath, &filesize, &content)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "Content not found or expired"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	// increment view count
	db.Exec(toSQL("UPDATE content SET view_count = view_count + 1 WHERE id = ?"), id)

	if contentType == "note" {
		c.JSON(http.StatusOK, gin.H{
			"type":    "note",
			"title":   title.String,
			"content": content.String,
		})
	} else if contentType == "file" {
		// check if file exists
		if _, err := os.Stat(filePath.String); os.IsNotExist(err) {
			c.JSON(http.StatusNotFound, gin.H{"error": "File not found"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"type":     "file",
			"filename": filename.String,
			"size":     filesize.Int64,
			"downloadUrl": fmt.Sprintf("/api/content/%s/download", id),
		})
	}
}

func handleGetStats(c *gin.Context) {
	id := c.Param("id")

	var viewCount int
	var createdAt, expiresAt time.Time

	err := db.QueryRow(toSQL(`
		SELECT view_count, created_at, expires_at
		FROM content
		WHERE id = ?
	`), id).Scan(&viewCount, &createdAt, &expiresAt)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "Content not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"viewCount": viewCount,
		"createdAt": createdAt.Format(time.RFC3339),
		"expiresAt": expiresAt.Format(time.RFC3339),
	})
}

func handleDownload(c *gin.Context) {
	id := c.Param("id")

	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID required"})
		return
	}

	var filename, filePath string
	var contentType string

	err := db.QueryRow(toSQL(`
		SELECT filename, filepath, type
		FROM content
		WHERE id = ? AND expires_at > datetime('now')
	`), id).Scan(&filename, &filePath, &contentType)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "Content not found or expired"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	if contentType != "file" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Content is not a file"})
		return
	}

	// check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, gin.H{"error": "File not found"})
		return
	}

	// increment view count
	db.Exec(toSQL("UPDATE content SET view_count = view_count + 1 WHERE id = ?"), id)

	// serve file for download
	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Transfer-Encoding", "binary")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	c.Header("Content-Type", "application/octet-stream")
	c.File(filePath)
}

func handleAdminList(c *gin.Context) {
	// Check admin secret
	adminSecret := os.Getenv("ADMIN_SECRET")
	if adminSecret == "" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Admin endpoint disabled"})
		return
	}

	providedSecret := c.GetHeader("X-Admin-Secret")
	if providedSecret != adminSecret {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	rows, err := db.Query(toSQL(`
		SELECT id, type, title, filename, filesize, created_at, expires_at, view_count 
		FROM content 
		ORDER BY created_at DESC
	`))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}
	defer rows.Close()

	var contents []gin.H
	for rows.Next() {
		var id, contentType string
		var title, filename sql.NullString
		var filesize sql.NullInt64
		var createdAt, expiresAt time.Time
		var viewCount int

		if err := rows.Scan(&id, &contentType, &title, &filename, &filesize, &createdAt, &expiresAt, &viewCount); err != nil {
			continue
		}

		content := gin.H{
			"id":         id,
			"type":       contentType,
			"created_at": createdAt.Format(time.RFC3339),
			"expires_at": expiresAt.Format(time.RFC3339),
			"view_count": viewCount,
		}

		if title.Valid {
			content["title"] = title.String
		}
		if filename.Valid {
			content["filename"] = filename.String
		}
		if filesize.Valid {
			content["filesize"] = filesize.Int64
		}

		contents = append(contents, content)
	}

	c.JSON(http.StatusOK, gin.H{
		"total":    len(contents),
		"contents": contents,
	})
}
