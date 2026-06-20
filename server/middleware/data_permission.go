package middleware

import (
	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/system"
	"github.com/flipped-aurora/gin-vue-admin/server/utils"
	"github.com/gin-gonic/gin"
)

func DataPermission() gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, err := utils.GetClaims(c)
		if err != nil {
			c.Next()
			return
		}
		userId := claims.BaseClaims.ID
		authorityId := claims.AuthorityId
		var user system.SysUser
		if err := global.GVA_DB.Preload("Authorities").Where("id = ?", userId).First(&user).Error; err != nil {
			c.Set("data_permission_user_id", userId)
			c.Set("data_permission_authority_ids", []uint{authorityId})
			c.Next()
			return
		}
		authorityIds := make([]uint, 0, len(user.Authorities)+1)
		authorityIds = append(authorityIds, authorityId)
		for _, auth := range user.Authorities {
			if auth.AuthorityId != authorityId {
				authorityIds = append(authorityIds, auth.AuthorityId)
			}
		}
		c.Set("data_permission_user_id", userId)
		c.Set("data_permission_authority_ids", authorityIds)
		c.Next()
	}
}
