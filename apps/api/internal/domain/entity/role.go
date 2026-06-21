package entity

import "time"

type Permission string

const (
	PermContentRead       Permission = "content:read"
	PermContentCreate     Permission = "content:create"
	PermContentUpdate     Permission = "content:update"
	PermContentDelete     Permission = "content:delete"
	PermContentPublish    Permission = "content:publish"
	PermContentUnpublish  Permission = "content:unpublish"
	PermMediaRead         Permission = "media:read"
	PermMediaUpload       Permission = "media:upload"
	PermMediaDelete       Permission = "media:delete"
	PermUsersManage       Permission = "users:manage"
	PermRolesManage       Permission = "roles:manage"
	PermAccessTokenManage Permission = "access_tokens:manage"
	PermContentTypesRead  Permission = "content_types:read"
)

var AllPermissions = []Permission{
	PermContentRead, PermContentCreate, PermContentUpdate, PermContentDelete,
	PermContentPublish, PermContentUnpublish,
	PermMediaRead, PermMediaUpload, PermMediaDelete,
	PermUsersManage, PermRolesManage, PermAccessTokenManage,
	PermContentTypesRead,
}

func IsValidPermission(p string) bool {
	for _, ap := range AllPermissions {
		if string(ap) == p {
			return true
		}
	}
	return false
}

type RoleEntity struct {
	ID          string    `bson:"_id,omitempty"   gorm:"column:gorm_id;primaryKey"             json:"-"`
	DocumentID  string    `bson:"documentId"      gorm:"column:document_id;uniqueIndex"      json:"documentId"`
	Name        string    `bson:"name"            gorm:"column:name"                         json:"name"`
	Slug        string    `bson:"slug"            gorm:"column:slug;uniqueIndex"             json:"slug"`
	Permissions []string  `bson:"permissions"     gorm:"column:permissions;serializer:json"   json:"permissions"`
	Level       int       `bson:"level"           gorm:"column:level"                        json:"level"`
	IsDefault   bool      `bson:"isDefault"       gorm:"column:is_default"                   json:"isDefault"`
	CreatedAt   time.Time `bson:"createdAt"       gorm:"column:created_at"                   json:"createdAt"`
	UpdatedAt   time.Time `bson:"updatedAt"       gorm:"column:updated_at"                   json:"updatedAt"`
}

func (RoleEntity) TableName() string {
	return "roles"
}

func AllPermissionStrings() []string {
	out := make([]string, len(AllPermissions))
	for i, p := range AllPermissions {
		out[i] = string(p)
	}
	return out
}

var DefaultRoles = []RoleEntity{
	{
		Name:        "Super Admin",
		Slug:        "super_admin",
		Permissions: AllPermissionStrings(),
		Level:       100,
		IsDefault:   true,
	},
	{
		Name: "Admin",
		Slug: "admin",
		Permissions: []string{
			"content:read", "content:create", "content:update", "content:delete",
			"content:publish", "content:unpublish",
			"media:read", "media:upload", "media:delete",
			"users:manage", "content_types:read",
		},
		Level:     80,
		IsDefault: true,
	},
	{
		Name: "Editor",
		Slug: "editor",
		Permissions: []string{
			"content:read", "content:create", "content:update", "content:delete",
			"content:publish", "content:unpublish",
			"media:read", "media:upload", "media:delete",
			"content_types:read",
		},
		Level:     60,
		IsDefault: true,
	},
	{
		Name: "Guest",
		Slug: "guest",
		Permissions: []string{
			"content:read", "media:read", "content_types:read",
		},
		Level:     20,
		IsDefault: true,
	},
}
