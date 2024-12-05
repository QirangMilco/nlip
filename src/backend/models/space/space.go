package space

import (
	"time"
)

type Space struct {
	ID            string            `json:"id"`
	Name          string            `json:"name"`
	Type          string            `json:"type"` // public 或 private
	OwnerID       string            `json:"ownerId"`
	MaxItems      int               `json:"maxItems"`
	RetentionDays int               `json:"retentionDays"`
	CreatedAt     time.Time         `json:"createdAt"`
	UpdatedAt     time.Time         `json:"updatedAt"`
	InvitedUsers  map[string]string `json:"invitedUsers"`
}

type CreateSpaceRequest struct {
	Name          string            `json:"name" validate:"required,min=2,max=50"`
	Type          string            `json:"type" validate:"required,oneof=public private"`
	MaxItems      int               `json:"maxItems" validate:"required,min=1"`
	RetentionDays int               `json:"retentionDays" validate:"required,min=1"`
	InvitedUsers  map[string]string `json:"invitedUsers" validate:"omitempty,dive,keys,required,endkeys,oneof=edit view"`
}

type UpdateSpaceRequest struct {
	Name          string            `json:"name" validate:"omitempty,min=2,max=50"`
	MaxItems      int               `json:"maxItems,omitempty"`
	RetentionDays int               `json:"retentionDays,omitempty"`
	InvitedUsers  map[string]string `json:"invitedUsers,omitempty" validate:"omitempty,dive,keys,required,endkeys,oneof=edit view"`
}

type SpaceResponse struct {
	Space *Space `json:"space"`
}

type ListSpacesResponse struct {
	Spaces []Space `json:"spaces"`
}

type InviteCollaboratorRequest struct {
	CollaboratorID string `json:"collaboratorId" validate:"required"`
	Permission     string `json:"permission" validate:"required,oneof=edit view"`
}

type RemoveCollaboratorRequest struct {
	CollaboratorID string `json:"collaboratorId" validate:"required"`
}

type UpdateCollaboratorPermissionsRequest struct {
	CollaboratorID string `json:"collaboratorId" validate:"required"`
	Permission     string `json:"permission" validate:"required,oneof=edit view"`
}

type UpdateSpaceSettingsRequest struct {
	Name          string `json:"name" validate:"omitempty,min=2,max=50"`
	MaxItems      int    `json:"maxItems" validate:"omitempty,min=1"`
	RetentionDays int    `json:"retentionDays" validate:"omitempty,min=1"`
	Visibility    string `json:"visibility" validate:"omitempty,oneof=public private"`
}
