package main

import (
	"fmt"
	"strings"

	"github.com/barasher/go-exiftool"
	"github.com/bogem/id3v2/v2"
)

type Metadata struct {
	FileType string `json:"file_type"`
	Artist   string `json:"artist"`
	Title    string `json:"title,omitempty"`
	Album    string `json:"album,omitempty"`
	Year     string `json:"year,omitempty"`
	Model    string `json:"model,omitempty"`
}

func ReadAudioMetadata(filePath string) (*Metadata, error) {
	tag, err := id3v2.Open(filePath, id3v2.Options{Parse: true})
	if err != nil {
		return nil, fmt.Errorf("failed to open mp3: %w", err)
	}
	defer tag.Close()

	return &Metadata{
		FileType: "audio",
		Artist:   tag.Artist(),
		Title:    tag.Title(),
		Album:    tag.Album(),
		Year:     tag.Year(),
	}, nil
}

func ReadImageMetadata(et *exiftool.Exiftool, filePath string) (*Metadata, error) {
	fileInfos := et.ExtractMetadata(filePath)
	if len(fileInfos) == 0 {
		return nil, fmt.Errorf("no metadata found for %s", filePath)
	}
	info := fileInfos[0]
	if info.Err != nil {
		return nil, fmt.Errorf("exiftool error: %w", info.Err)
	}

	meta := &Metadata{
		FileType: "image",
	}

	if artist, ok := info.Fields["Artist"].(string); ok {
		meta.Artist = strings.TrimSpace(artist)
	}
	if model, ok := info.Fields["Model"].(string); ok {
		meta.Model = strings.TrimSpace(model)
	}
	if desc, ok := info.Fields["ImageDescription"].(string); ok {
		meta.Title = strings.TrimSpace(desc)
	}

	return meta, nil
}
func WriteAudioMetadata(filePath string, meta *Metadata) error {
	tag, err := id3v2.Open(filePath, id3v2.Options{Parse: true})
	if err != nil {
		return fmt.Errorf("failed to open mp3 for writing: %w", err)
	}
	defer tag.Close()

	tag.SetVersion(3)
	tag.SetDefaultEncoding(id3v2.EncodingISO)
	if meta.Artist != "" {
		tag.SetArtist(meta.Artist)
	}
	if meta.Title != "" {
		tag.SetTitle(meta.Title)
	}
	if meta.Album != "" {
		tag.SetAlbum(meta.Album)
	}
	if meta.Year != "" {
		tag.SetYear(meta.Year)
	}
	if err := tag.Save(); err != nil {
		return fmt.Errorf("failed to save mp3 tags: %w", err)
	}
	return nil
}
func WriteImageMetadata(et *exiftool.Exiftool, filePath string, meta *Metadata) error {
	fileInfos := et.ExtractMetadata(filePath)
	if len(fileInfos) == 0 {
		return fmt.Errorf("no metadata found for %s", filePath)
	}
	fileInfo := fileInfos[0]
	fileInfo.Err = nil

	if meta.Artist != "" {
		fileInfo.SetString("Artist", meta.Artist)
	}
	if meta.Title != "" {
		fileInfo.SetString("ImageDescription", meta.Title)
	}
	et.WriteMetadata([]exiftool.FileMetadata{fileInfo})
	if fileInfo.Err != nil {
		return fmt.Errorf("exiftool failed to write metadata: %w", fileInfo.Err)
	}
	return nil
}
