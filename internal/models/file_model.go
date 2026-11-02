package models

import "time"

type File struct {
	ID           string    `json:"id"`
	StorageKey   string    `json:"storage_key"`
	OriginalName string    `json:"original_name"`
	ContentType  string    `json:"content_type"`
	SizeBytes    int64     `json:"size_bytes"`
	Kind         string    `json:"kind"`
	Visibility   string    `json:"visibility"`
	Url          string    `json:"url"`
	UploadedBy   string    `json:"uploaded_by"`
	CreatedAt    time.Time `json:"created_at"`
}
