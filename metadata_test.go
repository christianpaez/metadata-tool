package main

import (
	"os"
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
	if meta.Artist != "Test Artist" {
		t.Errorf("expected Artist 'Test Artist', got %q", meta.Artist)
	}
	if meta.Model != "Virtual Camera" {
		t.Errorf("expected Model 'Virtual Camera', got %q", meta.Model)
	}
}
