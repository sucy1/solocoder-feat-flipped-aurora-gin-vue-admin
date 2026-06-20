package system

import (
	"github.com/flipped-aurora/gin-vue-admin/server/middleware"
	"github.com/gin-gonic/gin"
)

type FieldPermissionRouter struct{}

func (s *FieldPermissionRouter) InitFieldPermissionRouter(Router *gin.RouterGroup) {
	fieldPermissionRouter := Router.Group("fieldPermission").Use(middleware.OperationRecord())
	fieldPermissionRouterWithoutRecord := Router.Group("fieldPermission")
	{
		fieldPermissionRouter.POST("createFieldPermission", fieldPermissionApi.CreateFieldPermission)
		fieldPermissionRouter.DELETE("deleteFieldPermission", fieldPermissionApi.DeleteFieldPermission)
		fieldPermissionRouter.PUT("updateFieldPermission", fieldPermissionApi.UpdateFieldPermission)
	}
	{
		fieldPermissionRouterWithoutRecord.GET("findFieldPermission", fieldPermissionApi.FindFieldPermission)
		fieldPermissionRouterWithoutRecord.GET("getFieldPermissionList", fieldPermissionApi.GetFieldPermissionList)
		fieldPermissionRouterWithoutRecord.GET("getFieldPermissionsForPath", fieldPermissionApi.GetFieldPermissionsForPath)
	}
}
