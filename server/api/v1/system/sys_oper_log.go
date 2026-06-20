package system

import (
	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/common/request"
	"github.com/flipped-aurora/gin-vue-admin/server/model/common/response"
	"github.com/flipped-aurora/gin-vue-admin/server/model/system"
	systemReq "github.com/flipped-aurora/gin-vue-admin/server/model/system/request"
	"github.com/flipped-aurora/gin-vue-admin/server/utils"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type OperLogApi struct{}

func (s *OperLogApi) CreateSysOperLog(c *gin.Context) {
	var sysOperLog system.SysOperLog
	err := c.ShouldBindJSON(&sysOperLog)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	err = operLogService.CreateSysOperLog(sysOperLog)
	if err != nil {
		global.GVA_LOG.Error("创建失败!", zap.Error(err))
		response.FailWithMessage("创建失败", c)
		return
	}
	response.OkWithMessage("创建成功", c)
}

func (s *OperLogApi) DeleteSysOperLog(c *gin.Context) {
	var sysOperLog system.SysOperLog
	err := c.ShouldBindJSON(&sysOperLog)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	err = operLogService.DeleteSysOperLog(sysOperLog)
	if err != nil {
		global.GVA_LOG.Error("删除失败!", zap.Error(err))
		response.FailWithMessage("删除失败", c)
		return
	}
	response.OkWithMessage("删除成功", c)
}

func (s *OperLogApi) DeleteSysOperLogByIds(c *gin.Context) {
	var IDS request.IdsReq
	err := c.ShouldBindJSON(&IDS)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	err = operLogService.DeleteSysOperLogByIds(IDS)
	if err != nil {
		global.GVA_LOG.Error("批量删除失败!", zap.Error(err))
		response.FailWithMessage("批量删除失败", c)
		return
	}
	response.OkWithMessage("批量删除成功", c)
}

func (s *OperLogApi) FindSysOperLog(c *gin.Context) {
	var sysOperLog system.SysOperLog
	err := c.ShouldBindQuery(&sysOperLog)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	err = utils.Verify(sysOperLog, utils.IdVerify)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	reSysOperLog, err := operLogService.GetSysOperLog(sysOperLog.ID)
	if err != nil {
		global.GVA_LOG.Error("查询失败!", zap.Error(err))
		response.FailWithMessage("查询失败", c)
		return
	}
	response.OkWithDetailed(gin.H{"reSysOperLog": reSysOperLog}, "查询成功", c)
}

func (s *OperLogApi) GetSysOperLogList(c *gin.Context) {
	var pageInfo systemReq.SysOperLogSearch
	err := c.ShouldBindQuery(&pageInfo)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	list, total, err := operLogService.GetSysOperLogInfoList(pageInfo)
	if err != nil {
		global.GVA_LOG.Error("获取失败!", zap.Error(err))
		response.FailWithMessage("获取失败", c)
		return
	}
	response.OkWithDetailed(response.PageResult{
		List:     list,
		Total:    total,
		Page:     pageInfo.Page,
		PageSize: pageInfo.PageSize,
	}, "获取成功", c)
}
