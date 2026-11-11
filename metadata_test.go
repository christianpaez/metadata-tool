package main

import (
	"os"
	"testing"

	"io"
	"path/filepath"

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
	if meta.Artist != "Test Artist" {
		t.Errorf("expected Artist 'Test Artist', got %q", meta.Artist)
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
		Artist: "Lead Photographer",
		Title:  "Cover Photo Description",
	}
	if err := WriteImageMetadata(et, tempImage, update); err != nil {
		t.Fatalf("failed to write image metadata: %v", err)
	}
	readBack, err := ReadImageMetadata(et, tempImage)
	if err != nil {
		t.Fatalf("failed to read back image metadata: %v", err)
	}
	if readBack.Artist != "Lead Photographer" {
		t.Errorf("expected Artist 'Lead Photographer', got %q", readBack.Artist)
	}
	if readBack.Title != "Cover Photo Description" {
		t.Errorf("expected Title 'Cover Photo Description', got %q", readBack.Title)
	}
	if readBack.Model != "Virtual Camera" {
		t.Errorf("expected camera Model 'Virtual Camera' to be preserved, got %q", readBack.Model)
	}
}
