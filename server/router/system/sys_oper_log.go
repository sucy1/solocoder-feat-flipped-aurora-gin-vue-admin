package system

import (
	"github.com/gin-gonic/gin"
)

type OperLogRouter struct{}

func (s *OperLogRouter) InitOperLogRouter(Router *gin.RouterGroup) {
	operLogRouter := Router.Group("sysOperLog")
	{
		operLogRouter.DELETE("deleteSysOperLog", operLogApi.DeleteSysOperLog)
		operLogRouter.DELETE("deleteSysOperLogByIds", operLogApi.DeleteSysOperLogByIds)
		operLogRouter.GET("findSysOperLog", operLogApi.FindSysOperLog)
		operLogRouter.GET("getSysOperLogList", operLogApi.GetSysOperLogList)
	}
}
