package clip

import (
    "time"
)

type Clip struct {
    ID          string    `json:"id"`
    SpaceID     string    `json:"spaceId"`
    ContentType string    `json:"contentType"`
    Content     string    `json:"content,omitempty"`
    FilePath    string    `json:"filePath,omitempty"`
    CreatedAt   time.Time `json:"createdAt"`
}

type UploadClipRequest struct {
    SpaceID     string `json:"spaceId" validate:"required"`
    ContentType string `json:"contentType" validate:"required"`
    Content     string `json:"content,omitempty"`
    File        []byte `json:"-"`
    FileName    string `json:"fileName,omitempty"`
}

type ClipResponse struct {
    Clip *Clip `json:"clip"`
}

type ListClipsResponse struct {
    Clips []Clip `json:"clips"`
} 