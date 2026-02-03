package models

import (
	"time"
)

// shortened url represents a shortened url record
type ShortenedURL struct {
	ID          int64      `db:"id" json:"id"`
	ShortCode   string     `db:"short_code" json:"shortCode"`
	OriginalURL string     `db:"original_url" json:"originalUrl"`
	CustomAlias *string    `db:"custom_alias" json:"customAlias,omitempty"`
	CreatedAt   time.Time  `db:"created_at" json:"createdAt"`
	ExpiresAt   *time.Time `db:"expires_at" json:"expiresAt,omitempty"`
	ClickCount  int        `db:"click_count" json:"clickCount"`
	DeletedAt   *time.Time `db:"deleted_at" json:"deletedAt,omitempty"`
}

// url click represents a click analytics record
type URLClick struct {
	ID         int64     `db:"id" json:"id"`
	URLID      int64     `db:"url_id" json:"urlId"`
	ClickedAt  time.Time `db:"clicked_at" json:"clickedAt"`
	IPAddress  string    `db:"ip_address" json:"ipAddress"`
	UserAgent  string    `db:"user_agent" json:"userAgent"`
	Referrer   string    `db:"referrer" json:"referrer"`
}

// shorten request represents url shortening request
type ShortenRequest struct {
	URL         string  `json:"url" binding:"required"`
	CustomAlias *string `json:"customAlias,omitempty"`
	ExpiresIn   *int    `json:"expiresIn,omitempty"` // days
}

// shorten response represents api response
type ShortenResponse struct {
	ID          int64      `json:"id"`
	ShortCode   string     `json:"shortCode"`
	ShortURL    string     `json:"shortUrl"`
	OriginalURL string     `json:"originalUrl"`
	CreatedAt   time.Time  `json:"createdAt"`
	ExpiresAt   *time.Time `json:"expiresAt,omitempty"`
}

// url stats response represents analytics data
type URLStatsResponse struct {
	ShortCode   string    `json:"shortCode"`
	OriginalURL string    `json:"originalUrl"`
	ClickCount  int       `json:"clickCount"`
	CreatedAt   time.Time `json:"createdAt"`
	ExpiresAt   *time.Time `json:"expiresAt,omitempty"`
	RecentClicks []URLClick `json:"recentClicks,omitempty"`
}
