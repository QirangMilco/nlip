package clip

import (
    "time"
)

type Clip struct {
    ID          string    `json:"id"`
    ClipID      string    `json:"clipId"`
    SpaceID     string    `json:"spaceId"`
    ContentType string    `json:"contentType"`
    Content     string   `json:"content,omitempty"`
    FilePath    string   `json:"filePath,omitempty"`
    Creator     *Creator  `json:"creator,omitempty"`
    CreatedAt   time.Time `json:"createdAt"`
    UpdatedAt   time.Time `json:"updatedAt"`
}

type Creator struct {
    ID       string `json:"id"`
    Username string `json:"username"`
}

type UploadClipRequest struct {
    SpaceID     string `json:"spaceId"`
    Content     string `json:"content"`
    ContentType string `json:"contentType"`
    Creator     string `json:"creator,omitempty"`
    File        []byte `json:"-"`
    FileName    string `json:"-"`
}

type ClipResponse struct {
    Clip *Clip `json:"clip"`
}

type ListClipsResponse struct {
    Clips []Clip `json:"clips"`
}

type UpdateClipRequest struct {
    Content string `json:"content" validate:"required"`
} 