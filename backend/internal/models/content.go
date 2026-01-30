package models

import (
	"time"
)

// content represents a shared item (file or note)
type Content struct {
	ID        string    `db:"id" json:"id"`
	Type      string    `db:"type" json:"type"`
	Title     *string   `db:"title" json:"title,omitempty"`
	Filename  *string   `db:"filename" json:"filename,omitempty"`
	Filepath  *string   `db:"filepath" json:"filepath,omitempty"`
	Filesize  *int64    `db:"filesize" json:"filesize,omitempty"`
	Content   *string   `db:"content" json:"content,omitempty"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	ExpiresAt time.Time `db:"expires_at" json:"expires_at"`
	ViewCount int       `db:"view_count" json:"view_count"`
	DeletedAt *time.Time `db:"deleted_at" json:"deleted_at,omitempty"`
}

// content type constants
const (
	ContentTypeFile = "file"
	ContentTypeNote = "note"
)

// upload request represents file upload data
type UploadRequest struct {
	File     []byte
	Filename string
	Size     int64
}

// note request represents note creation data
type NoteRequest struct {
	Title   string `json:"title"`
	Content string `json:"content" binding:"required"`
}

// content response represents api response
type ContentResponse struct {
	ID          string  `json:"id"`
	Type        string  `json:"type"`
	Title       *string `json:"title,omitempty"`
	Filename    *string `json:"filename,omitempty"`
	Size        *int64  `json:"size,omitempty"`
	Content     *string `json:"content,omitempty"`
	DownloadURL *string `json:"downloadUrl,omitempty"`
	ExpiresAt   string  `json:"expiresAt"`
}

// stats response represents content statistics
type StatsResponse struct {
	ViewCount int    `json:"viewCount"`
	CreatedAt string `json:"createdAt"`
	ExpiresAt string `json:"expiresAt"`
}
