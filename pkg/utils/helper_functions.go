package CustomUtils

import (
	"path/filepath"
	"strings"

	"github.com/google/uuid"
)

// Generic helper that safely dereferences a pointer
func Deref[T any](p *T) T {
	var zero T
	if p == nil {
		return zero
	}
	return *p
}

func DerefSlice[T any](items []*T) []T {
	result := make([]T, 0, len(items))

	for _, item := range items {
		if item != nil {
			result = append(result, *item)
		}
	}

	return result
}

func GenerateObjectKey(
	visibility string,
	kind string,
	extension string,
) string {
	folder := strings.ToLower(visibility) + "/"

	if kind == "IMAGE" {
		folder += "images/"
	} else {
		folder += "documents/"
	}

	extension = strings.ToLower(filepath.Ext(extension))

	return folder + uuid.NewString() + extension
}

// Helper function for getting prefixed column names for join operations in sql queries
func PrefixColumns(alias, cols string) string {
	parts := strings.Split(cols, ",")
	for i, p := range parts {
		parts[i] = alias + "." + strings.TrimSpace(p)
	}
	return strings.Join(parts, ", ")
}
