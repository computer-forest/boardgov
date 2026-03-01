package validator

import (
	"errors"
	"net/http"
	"strings"
)

var (
	allowedMagicBytes = map[string][]byte{
		"application/pdf": {0x25, 0x50, 0x44, 0x46},
		"application/vnd.openxmlformats-officedocument.wordprocessingml.document": {0x50, 0x4B, 0x03, 0x04},
	}
)

func ValidateFileMIME(header []byte, declared string) error {
	magic, ok := allowedMagicBytes[declared]
	if !ok {
		return errors.New("file type not allowed")
	}
	if len(header) < len(magic) {
		return errors.New("file too small to validate")
	}
	for i, b := range magic {
		if header[i] != b {
			return errors.New("file signature mismatch: possible file type spoofing")
		}
	}
	return nil
}

func DetectContentType(data []byte) string {
	return http.DetectContentType(data)
}

func SanitizeFilename(name string) string {
	name = strings.ReplaceAll(name, "/", "")
	name = strings.ReplaceAll(name, "..", "")
	name = strings.TrimSpace(name)
	return name
}
