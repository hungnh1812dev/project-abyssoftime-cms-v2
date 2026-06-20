package middleware

import (
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"

	"project-abyssoftime-cms-v2/api/internal/domain/entity"
)

type RoleCache struct {
	mu    sync.RWMutex
	roles map[string]*entity.RoleEntity
}

func NewRoleCache() *RoleCache {
	return &RoleCache{roles: make(map[string]*entity.RoleEntity)}
}

func (c *RoleCache) Load(roles []*entity.RoleEntity) {
	c.mu.Lock()
	defer c.mu.Unlock()
	m := make(map[string]*entity.RoleEntity, len(roles))
	for _, r := range roles {
		m[r.Slug] = r
	}
	c.roles = m
}

func (c *RoleCache) HasPermission(roleSlug, permission string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	r, ok := c.roles[roleSlug]
	if !ok {
		return false
	}
	for _, p := range r.Permissions {
		if p == permission {
			return true
		}
	}
	return false
}

func (c *RoleCache) GetLevel(roleSlug string) int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	r, ok := c.roles[roleSlug]
	if !ok {
		return 0
	}
	return r.Level
}

func (c *RoleCache) Get(roleSlug string) *entity.RoleEntity {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.roles[roleSlug]
}

func GinRequirePermission(cache *RoleCache, permission string) gin.HandlerFunc {
	return func(c *gin.Context) {
		roleSlug := c.GetString("role")
		if !cache.HasPermission(roleSlug, permission) {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "forbidden"})
			return
		}
		c.Next()
	}
}
