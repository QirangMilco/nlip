package space

import (
	"time"
)

type Space struct {
	ID            string             `json:"id"`
	Name          string             `json:"name"`
	Type          string             `json:"type"` // public 或 private
	OwnerID       string             `json:"ownerId"`
	MaxItems      int                `json:"maxItems"`
	RetentionDays int                `json:"retentionDays"`
	CreatedAt     time.Time          `json:"createdAt"`
	UpdatedAt     time.Time          `json:"updatedAt"`
	Collaborators []CollaboratorInfo `json:"collaborators"`
	CollaboratorsMap map[string]string `json:"collaboratorsMap"`
}

type CreateSpaceRequest struct {
	Name          string             `json:"name" validate:"required,min=2,max=50"`
	Type          string             `json:"type" validate:"omitempty,oneof=public private"`
	MaxItems      int                `json:"maxItems" validate:"required,min=1"`
	RetentionDays int                `json:"retentionDays" validate:"required,min=1"`
	Collaborators []CollaboratorInfo `json:"collaborators" validate:"omitempty,dive,keys,required,endkeys,oneof=edit view"`
}

type UpdateSpaceRequest struct {
	Name          string             `json:"name" validate:"omitempty,min=2,max=50"`
	MaxItems      int                `json:"maxItems,omitempty"`
	RetentionDays int                `json:"retentionDays,omitempty"`
	Collaborators []CollaboratorInfo `json:"collaborators,omitempty" validate:"omitempty,dive,keys,required,endkeys,oneof=edit view"`
}

type SpaceResponse struct {
	Space *Space `json:"space"`
}

type ListSpacesResponse struct {
	Spaces []Space `json:"spaces"`
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

// InviteCollaboratorRequest 邀请协作者请求
type InviteCollaboratorRequest struct {
	Permission string `json:"permission" validate:"required,oneof=edit view"`
	Email      string `json:"email" validate:"required,email"`
}

// InviteCollaboratorResponse 邀请协作者响应
type InviteCollaboratorResponse struct {
	InviteLink string `json:"inviteLink"`
}

// VerifyInviteTokenResponse 验证邀请令牌响应
type VerifyInviteTokenResponse struct {
	SpaceID           string `json:"spaceId"`
	SpaceName         string `json:"spaceName"`
	InviterName       string `json:"inviterName"`
	Permission        string `json:"permission"`
	IsCollaborator    bool   `json:"isCollaborator"`
}

// ValidateInviteRequest 验证邀请令牌请求
type ValidateInviteRequest struct {
	Token string `json:"token" validate:"required"`
}

// AcceptInviteRequest 接受邀请请求
type AcceptInviteRequest struct {
	Token string `json:"token" validate:"required"`
}

// CollaboratorInfo 协作者信息
type CollaboratorInfo struct {
	ID         string `json:"id"`
	Username   string `json:"username"`
	Permission string `json:"permission"`
}

// ListCollaboratorsResponse 获取协作者列表的响应
type ListCollaboratorsResponse struct {
	Collaborators []CollaboratorInfo `json:"collaborators"`
}
