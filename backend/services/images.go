package services

import (
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type ImageService struct {
	uploadDir string
}

func NewImageService(uploadDir string) *ImageService {
	if _, err := os.Stat(uploadDir); os.IsNotExist(err) {
		os.MkdirAll(uploadDir, 0o755)
	}

	return &ImageService{
		uploadDir: uploadDir,
	}
}

func (s *ImageService) ValidateImage(handler *multipart.FileHeader) (bool, error) {
	if handler.Size > 5<<20 {
		return false, nil
	}

	// Check file type
	contentType := handler.Header.Get("Content-Type")
	if contentType != "image/jpeg" && contentType != "image/png" && contentType != "image/gif" {
		return false, nil
	}

	// Check file extension
	ext := filepath.Ext(handler.Filename)
	ext = strings.ToLower(ext)
	if ext != ".jpg" && ext != ".jpeg" && ext != ".png" && ext != ".gif" {
		return false, nil
	}

	return true, nil
}

func (s *ImageService) SaveImage(file multipart.File, handler *multipart.FileHeader, subdir string) (string, error) {
	dirPath := filepath.Join(s.uploadDir, subdir)
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		os.MkdirAll(dirPath, 0o755)
	}

	ext := filepath.Ext(handler.Filename)
	newFileName := time.Now().Format("20060102150405") + "_" + GenerateRandomString(8) + ext
	filePath := filepath.Join(dirPath, newFileName)

	dst, err := os.Create(filePath)
	if err != nil {
		return "", err
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		return "", err
	}

	return filepath.Join(subdir, newFileName), nil
}

func (s *ImageService) GetImagePath(relativePath string) string {
	return filepath.Join(s.uploadDir, relativePath)
}

func GenerateRandomString(length int) string {
	return "random"
}
