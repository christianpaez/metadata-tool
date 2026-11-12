package main

import (
	"bytes"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/barasher/go-exiftool"
)

func TestReadAudioMetadata(t *testing.T) {
	audioPath := "testdata/synthetic.mp3"
	if _, err := os.Stat(audioPath); os.IsNotExist(err) {
		t.Fatalf("test fixture not found: %s", audioPath)
	}
	meta, err := ReadAudioMetadata(audioPath)
	if err != nil {
		t.Fatalf("unexpected error reading audio: %v", err)
	}
	if meta.FileType != "audio" {
		t.Errorf("expected FileType 'audio', got %q", meta.FileType)
	}
	if meta.Artist != "Test Artist" {
		t.Errorf("expected Artist 'Test Artist', got %q", meta.Artist)
	}
	if meta.Title != "Synthetic Track" {
		t.Errorf("expected Title 'Synthetic Track', got %q", meta.Title)
	}
	if meta.Album != "Test Suite Album" {
		t.Errorf("expected Album 'Test Suite Album', got %q", meta.Album)
	}
}
func TestReadImageMetadata(t *testing.T) {
	imagePath := "testdata/synthetic.jpg"
	if _, err := os.Stat(imagePath); os.IsNotExist(err) {
		t.Fatalf("test fixture not found: %s", imagePath)
	}
	et, err := exiftool.NewExiftool()
	if err != nil {
		t.Fatalf("failed to initialize exiftool: %v", err)
	}
	defer et.Close()
	meta, err := ReadImageMetadata(et, imagePath)
	if err != nil {
		t.Fatalf("unexpected error reading image: %v", err)
	}
	if meta.FileType != "image" {
		t.Errorf("expected FileType 'image', got %q", meta.FileType)
	}
	if meta.Model != "Virtual Camera" {
		t.Errorf("expected Model 'Virtual Camera', got %q", meta.Model)
	}
}

func copyFixtureToTemp(t *testing.T, src, dst string) {
	t.Helper()
	in, err := os.Open(src)
	if err != nil {
		t.Fatalf("failed to open fixture %s: %v", src, err)
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		t.Fatalf("failed to create temp file %s: %v", dst, err)
	}
	defer out.Close()
	if _, err := io.Copy(out, in); err != nil {
		t.Fatalf("failed to copy fixture: %v", err)
	}
}
func TestWriteAudioMetadata(t *testing.T) {
	tempDir := t.TempDir()
	tempAudio := filepath.Join(tempDir, "test.mp3")
	copyFixtureToTemp(t, "testdata/synthetic.mp3", tempAudio)
	update := &Metadata{
		Artist: "Slash",
		Title:  "World on Fire",
		Album:  "Apocalyptic Love",
		Year:   "2012",
	}
	if err := WriteAudioMetadata(tempAudio, update); err != nil {
		t.Fatalf("failed to write audio metadata: %v", err)
	}
	readBack, err := ReadAudioMetadata(tempAudio)
	if err != nil {
		t.Fatalf("failed to read back audio metadata: %v", err)
	}
	if readBack.Artist != "Slash" {
		t.Errorf("expected Artist 'Slash', got %q", readBack.Artist)
	}
	if readBack.Title != "World on Fire" {
		t.Errorf("expected Title 'World on Fire', got %q", readBack.Title)
	}
	if readBack.Album != "Apocalyptic Love" {
		t.Errorf("expected Album 'Apocalyptic Love', got %q", readBack.Album)
	}
	if readBack.Year != "2012" {
		t.Errorf("expected Year '2012', got %q", readBack.Year)
	}
}
func TestWriteImageMetadata(t *testing.T) {
	tempDir := t.TempDir()
	tempImage := filepath.Join(tempDir, "test.jpg")
	copyFixtureToTemp(t, "testdata/synthetic.jpg", tempImage)
	et, err := exiftool.NewExiftool()
	if err != nil {
		t.Fatalf("failed to initialize exiftool: %v", err)
	}
	defer et.Close()
	update := &Metadata{
		Copyright:   "© 2026 Test Photographer",
		UserComment: "An updated portrait",
	}
	if err := WriteImageMetadata(et, tempImage, update); err != nil {
		t.Fatalf("failed to write image metadata: %v", err)
	}
	readBack, err := ReadImageMetadata(et, tempImage)
	if err != nil {
		t.Fatalf("failed to read back image metadata: %v", err)
	}
	if readBack.Copyright != "© 2026 Test Photographer" {
		t.Errorf("expected Copyright '© 2026 Test Photographer', got %q", readBack.Copyright)
	}
	if readBack.UserComment != "An updated portrait" {
		t.Errorf("expected UserComment 'An updated portrait', got %q", readBack.UserComment)
	}
	if readBack.Model != "Virtual Camera" {
		t.Errorf("expected camera Model 'Virtual Camera' to be preserved, got %q", readBack.Model)
	}
}

func TestHandleEdit_UploadsAndModifiesImage(t *testing.T) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	writer.WriteField("copyright", "© 2026 Some Photographer")
	writer.WriteField("user_comment", "A beautiful test image")
	part, err := writer.CreateFormFile("file", "photo.jpg")
	if err != nil {
		t.Fatalf("failed to create form file: %v", err)
	}
	fixture, err := os.Open("testdata/synthetic.jpg")
	if err != nil {
		t.Fatalf("failed to open fixture: %v", err)
	}
	defer fixture.Close()
	io.Copy(part, fixture)
	writer.Close()
	req := httptest.NewRequest(http.MethodPost, "/edit", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	rr := httptest.NewRecorder()
	handleEdit(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", rr.Code, rr.Body.String())
	}
	disp := rr.Header().Get("Content-Disposition")
	if disp != `attachment; filename="photo.jpg"` {
		t.Errorf("unexpected Content-Disposition: %q", disp)
	}
	downloadedFile := filepath.Join(t.TempDir(), "downloaded.jpg")
	os.WriteFile(downloadedFile, rr.Body.Bytes(), 0644)
	et, err := exiftool.NewExiftool()
	if err != nil {
		t.Fatalf("failed to init exiftool: %v", err)
	}
	defer et.Close()
	meta, err := ReadImageMetadata(et, downloadedFile)
	if err != nil {
		t.Fatalf("failed to read back image metadata: %v", err)
	}
	if meta.Copyright != "© 2026 Some Photographer" {
		t.Errorf("expected Copyright '© 2026 Some Photographer', got %q", meta.Copyright)
	}
	if meta.UserComment != "A beautiful test image" {
		t.Errorf("expected UserComment 'A beautiful test image', got %q", meta.UserComment)
	}
	if meta.Model != "Virtual Camera" {
		t.Errorf("expected camera Model 'Virtual Camera' to be preserved, got %q", meta.Model)
	}
}

func TestHandleShowMedadata(t *testing.T) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", "photo.jpg")
	if err != nil {
		t.Fatalf("failed to create form file: %v", err)
	}
	fixture, err := os.Open("testdata/synthetic.jpg")
	if err != nil {
		t.Fatalf("failed to open fixture: %v", err)
	}
	defer fixture.Close()
	io.Copy(part, fixture)
	writer.Close()
	req := httptest.NewRequest(http.MethodPost, "/show", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	rr := httptest.NewRecorder()
	showHandler(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", rr.Code, rr.Body.String())
	}
	respBody := rr.Body.String()

	if !strings.Contains(respBody, "File Type: image") {
		t.Errorf("expected body to contain 'File Type: image'")
	}
	if !strings.Contains(respBody, "Camera Model: Virtual Camera") {
		t.Errorf("expected body to contain 'Virtual Camera', got %q", respBody)
	}
}
