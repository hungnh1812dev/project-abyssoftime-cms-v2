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

func (cache *RoleCache) Load(roles []*entity.RoleEntity) {
	cache.mu.Lock()
	defer cache.mu.Unlock()
	mapping := make(map[string]*entity.RoleEntity, len(roles))
	for _, roleEntity := range roles {
		mapping[roleEntity.Slug] = roleEntity
	}
	cache.roles = mapping
}

func (cache *RoleCache) HasPermission(roleSlug, permission string) bool {
	cache.mu.RLock()
	defer cache.mu.RUnlock()
	roleEntity, ok := cache.roles[roleSlug]
	if !ok {
		return false
	}
	for _, perm := range roleEntity.Permissions {
		if perm == permission {
			return true
		}
	}
	return false
}

func (cache *RoleCache) GetLevel(roleSlug string) int {
	cache.mu.RLock()
	defer cache.mu.RUnlock()
	roleEntity, ok := cache.roles[roleSlug]
	if !ok {
		return 0
	}
	return roleEntity.Level
}

func (cache *RoleCache) Get(roleSlug string) *entity.RoleEntity {
	cache.mu.RLock()
	defer cache.mu.RUnlock()
	return cache.roles[roleSlug]
}

func GinRequirePermission(cache *RoleCache, permission string) gin.HandlerFunc {
	return func(ginCtx *gin.Context) {
		roleSlug := ginCtx.GetString("role")
		if !cache.HasPermission(roleSlug, permission) {
			ginCtx.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "forbidden"})
			return
		}
		ginCtx.Next()
	}
}
