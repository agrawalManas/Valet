package repository

import (
	"agrawalManas/Valet/internal/models"
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func CreateFiles(pool *pgxpool.Pool, filesToUpload []models.File) ([]models.File, error) {
	var ctx context.Context
	var cancel context.CancelFunc

	ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if len(filesToUpload) == 0 {
		return nil, fmt.Errorf("no files provided")
	}

	tx, err := pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	var files []models.File
	query := `
		INSERT INTO files (
			storage_key,
			original_name,
			content_type,
			size_bytes,
			kind,
			visibility,
			url,
			uploaded_by
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, storage_key, original_name, content_type, size_bytes, kind, visibility, url, uploaded_by, created_at
	`

	for _, file := range filesToUpload {
		var addedFile models.File

		err := pool.QueryRow(
			ctx,
			query,
			file.StorageKey,
			file.OriginalName,
			file.ContentType,
			file.SizeBytes,
			file.Kind,
			file.Visibility,
			file.Url,
			file.UploadedBy,
		).Scan(
			&addedFile.ID,
			&addedFile.StorageKey,
			&addedFile.OriginalName,
			&addedFile.ContentType,
			&addedFile.SizeBytes,
			&addedFile.Kind,
			&addedFile.Visibility,
			&addedFile.Url,
			&addedFile.UploadedBy,
			&addedFile.CreatedAt,
		)

		if err != nil {
			return nil, fmt.Errorf(
				"failed to insert metadata for %s: %w",
				file.OriginalName,
				err,
			)
		}
		files = append(files, addedFile)
	}

	return files, nil
}
