package service

import (
	"agrawalManas/Valet/internal/constants"
	"agrawalManas/Valet/internal/models"
	"agrawalManas/Valet/internal/repository"
	"agrawalManas/Valet/internal/storage"
	CustomUtils "agrawalManas/Valet/pkg/utils"
	"context"
	"fmt"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	maxFileSize = 10 << 20 // 10 MB per file
	maxFiles    = 10
)

var allowedContentTypes = map[string]bool{
	"image/jpeg":      true,
	"image/png":       true,
	"image/webp":      true,
	"image/gif":       true,
	"image/svg":       true,
	"application/pdf": true,
}

func UploadFiles(
	ctx context.Context,
	files []*multipart.FileHeader,
	r2Client *storage.R2Client,
	userID string,
	visibility string,
	pool *pgxpool.Pool,
) ([]models.File, int, error) {

	// 1. Validate request.
	if len(files) == 0 {
		return nil, http.StatusBadRequest, fmt.Errorf(constants.REQUIRE_AT_LEAST_ONE_FILE)
	}

	if len(files) > maxFiles {
		return nil, http.StatusBadRequest, fmt.Errorf("maximum %d files allowed", maxFiles)
	}

	// 2. Validate every file before uploading anything.
	for _, file := range files {
		if file == nil {
			return nil, http.StatusBadRequest, fmt.Errorf("invalid file")
		}

		if file.Size <= 0 || file.Size > maxFileSize {
			return nil, http.StatusBadRequest, fmt.Errorf(
				"file %s must be between 1 byte and 10 MB",
				file.Filename,
			)
		}

		contentType, err := detectFileContentType(file)
		if err != nil {
			return nil, http.StatusBadRequest, fmt.Errorf(
				"failed to detect content type for %s: %w",
				file.Filename,
				err,
			)
		}

		if !allowedContentTypes[contentType] {
			return nil, http.StatusBadRequest, fmt.Errorf(
				"unsupported file type %s for %s",
				contentType,
				file.Filename,
			)
		}

		if visibility == constants.FILE_VISIBILITY_PUBLIC && !strings.HasPrefix(contentType, "image/") {
			return nil, http.StatusBadRequest, fmt.Errorf(
				"public uploads only support images: %s",
				file.Filename,
			)
		}
	}

	// 3. Upload files to R2
	uploadedFiles := make([]*models.File, 0, len(files))
	uploadedKeys := make([]string, 0, len(files))

	cleanup := func() {
		cleanupCtx, cancel := context.WithTimeout(
			context.Background(),
			15*time.Second,
		)
		defer cancel()

		for _, key := range uploadedKeys {
			if err := r2Client.Delete(cleanupCtx, key); err != nil {
				fmt.Printf("failed to delete R2 object %s: %v\n", key, err)
			}
		}
	}

	for _, file := range files {
		src, err := file.Open()
		if err != nil {
			cleanup()
			return nil, http.StatusInternalServerError, fmt.Errorf("failed to open %s: %w", file.Filename, err)
		}

		contentType, err := detectFileContentType(file)
		if err != nil {
			src.Close()
			cleanup()
			return nil, http.StatusInternalServerError, err
		}

		kind := constants.FILE_KIND_FILE
		if strings.HasPrefix(contentType, "image/") {
			kind = constants.FILE_KIND_IMAGE
		}

		extension := filepath.Ext(file.Filename)
		key := CustomUtils.GenerateObjectKey(visibility, kind, extension)

		err = r2Client.UploadObject(
			ctx,
			key,
			src,
			file.Size,
			contentType,
		)

		src.Close()

		if err != nil {
			cleanup()
			return nil, http.StatusInternalServerError, fmt.Errorf(
				"failed to upload %s to R2: %w",
				file.Filename,
				err,
			)
		}

		uploadedKeys = append(uploadedKeys, key)

		uploadedFiles = append(uploadedFiles, &models.File{
			StorageKey:   key,
			OriginalName: filepath.Base(file.Filename),
			ContentType:  contentType,
			SizeBytes:    file.Size,
			Kind:         kind,
			Visibility:   visibility,
			UploadedBy:   userID,
		})
	}

	// 4. Generate URLs in the same order as the uploaded files
	for _, file := range uploadedFiles {
		url, err := r2Client.GetObjectURL(
			ctx,
			file.StorageKey,
			file.Visibility,
		)

		if err != nil {
			cleanup()
			return nil, http.StatusInternalServerError, fmt.Errorf(
				"failed to generate URL for %s: %w",
				file.OriginalName,
				err,
			)
		}

		file.Url = url
	}

	// 5. Insert all file metadata in one DB transaction
	dbFiles, err := repository.CreateFiles(pool, CustomUtils.DerefSlice(uploadedFiles))
	if err != nil {
		cleanup()
		return nil, http.StatusInternalServerError, fmt.Errorf("failed to save file metadata: %w", err)
	}

	return dbFiles, http.StatusOK, nil
}

// detectFileContentType checks the actual file contents instead of trusting
// the Content-Type header supplied by the client.
func detectFileContentType(
	file *multipart.FileHeader,
) (string, error) {
	src, err := file.Open()
	if err != nil {
		return "", err
	}
	defer src.Close()

	buffer := make([]byte, 512)
	n, err := src.Read(buffer)
	if err != nil && n == 0 {
		return "", err
	}

	return http.DetectContentType(buffer[:n]), nil
}
