package system

import (
	"github.com/flipped-aurora/gin-vue-admin/server/middleware"
	"github.com/gin-gonic/gin"
)

type DataPermissionRouter struct{}

func (s *DataPermissionRouter) InitDataPermissionRouter(Router *gin.RouterGroup) {
	dataPermissionRouter := Router.Group("dataPermission").Use(middleware.OperationRecord())
	dataPermissionRouterWithoutRecord := Router.Group("dataPermission")
	{
		dataPermissionRouter.POST("createDataPermission", dataPermissionApi.CreateDataPermission)
		dataPermissionRouter.DELETE("deleteDataPermission", dataPermissionApi.DeleteDataPermission)
		dataPermissionRouter.PUT("updateDataPermission", dataPermissionApi.UpdateDataPermission)
	}
	{
		dataPermissionRouterWithoutRecord.GET("findDataPermission", dataPermissionApi.FindDataPermission)
		dataPermissionRouterWithoutRecord.GET("getDataPermissionList", dataPermissionApi.GetDataPermissionList)
	}
}
