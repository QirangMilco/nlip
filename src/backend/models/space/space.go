package space

import (
    "time"
)

type Space struct {
    ID            string    `json:"id"`
    Name          string    `json:"name"`
    Type          string    `json:"type"` // public 或 private
    OwnerID       string    `json:"ownerId"`
    MaxItems      int       `json:"maxItems"`
    RetentionDays int       `json:"retentionDays"`
    CreatedAt     time.Time `json:"createdAt"`
}

type CreateSpaceRequest struct {
    Name          string `json:"name" validate:"required,min=2,max=50"`
    Type          string `json:"type" validate:"required,oneof=public private"`
    MaxItems      int    `json:"maxItems" validate:"required,min=1,max=20"`
    RetentionDays int    `json:"retentionDays" validate:"required,min=1,max=7"`
}

type UpdateSpaceRequest struct {
    Name          string `json:"name" validate:"omitempty,min=2,max=50"`
    MaxItems      int    `json:"maxItems" validate:"omitempty,min=1,max=20"`
    RetentionDays int    `json:"retentionDays" validate:"omitempty,min=1,max=7"`
}

type SpaceResponse struct {
    Space *Space `json:"space"`
}

type ListSpacesResponse struct {
    Spaces []Space `json:"spaces"`
} 