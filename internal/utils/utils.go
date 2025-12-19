package utils

import (
	"crypto/rand"
	"encoding/hex"
	"strings"
	"time"
	"unicode"
)

func Slugify(name string) string {
	name = strings.ToLower(strings.TrimSpace(name))

	var b strings.Builder
	for _, r := range name {
		if unicode.IsLetter(r) || unicode.IsNumber(r) {
			b.WriteRune(r)
		} else {
			b.WriteRune('-')
		}
	}

	slug := b.String()
	slug = strings.ReplaceAll(slug, "--", "-")
	slug = strings.Trim(slug, "-")
	return slug
}

func GenerateAPIKey() (string, error) {
	randomBytes := make([]byte, 16)
	_, err := rand.Read(randomBytes)
	if err != nil {
		return "", err
	}

	return "bkps_" + hex.EncodeToString(randomBytes), nil
}

func GenerateID() string {
	bytes := make([]byte, 16)
	_, _ = rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

func GetShortDate(currTime time.Time) string {
	shortDateLayout := "Mon 02-Jan-06 2006"
	return currTime.Format(shortDateLayout)
}

func ToInt64Ptr(v int64) *int64 {
	return &v
}
