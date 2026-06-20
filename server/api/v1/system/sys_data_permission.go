package system

import (
	"context"
	"strconv"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/common/response"
	"github.com/flipped-aurora/gin-vue-admin/server/model/system"
	systemReq "github.com/flipped-aurora/gin-vue-admin/server/model/system/request"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type DataPermissionApi struct{}

func (d *DataPermissionApi) CreateDataPermission(c *gin.Context) {
	var dp systemReq.SysDataPermissionCreate
	err := c.ShouldBindJSON(&dp)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	dataPermission := system.SysDataPermission{
		AuthorityID:      dp.AuthorityID,
		Level:            dp.Level,
		CustomConditions: dp.CustomConditions,
	}
	err = dataPermissionService.CreateDataPermission(context.Background(), dataPermission)
	if err != nil {
		global.GVA_LOG.Error("创建失败!", zap.Error(err))
		response.FailWithMessage("创建失败:"+err.Error(), c)
		return
	}
	response.OkWithMessage("创建成功", c)
}

func (d *DataPermissionApi) UpdateDataPermission(c *gin.Context) {
	var dp systemReq.SysDataPermissionUpdate
	err := c.ShouldBindJSON(&dp)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	dataPermission := system.SysDataPermission{
		Level:            dp.Level,
		CustomConditions: dp.CustomConditions,
	}
	dataPermission.ID = dp.ID
	err = dataPermissionService.UpdateDataPermission(context.Background(), dataPermission)
	if err != nil {
		global.GVA_LOG.Error("更新失败!", zap.Error(err))
		response.FailWithMessage("更新失败:"+err.Error(), c)
		return
	}
	response.OkWithMessage("更新成功", c)
}

func (d *DataPermissionApi) DeleteDataPermission(c *gin.Context) {
	idStr := c.Query("ID")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		response.FailWithMessage("无效的ID参数", c)
		return
	}
	err = dataPermissionService.DeleteDataPermission(context.Background(), uint(id))
	if err != nil {
		global.GVA_LOG.Error("删除失败!", zap.Error(err))
		response.FailWithMessage("删除失败:"+err.Error(), c)
		return
	}
	response.OkWithMessage("删除成功", c)
}

func (d *DataPermissionApi) FindDataPermission(c *gin.Context) {
	idStr := c.Query("ID")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		response.FailWithMessage("无效的ID参数", c)
		return
	}
	reDp, err := dataPermissionService.GetDataPermission(context.Background(), uint(id))
	if err != nil {
		global.GVA_LOG.Error("查询失败!", zap.Error(err))
		response.FailWithMessage("查询失败:"+err.Error(), c)
		return
	}
	response.OkWithDetailed(reDp, "查询成功", c)
}

func (d *DataPermissionApi) GetDataPermissionList(c *gin.Context) {
	var pageInfo systemReq.SysDataPermissionSearch
	err := c.ShouldBindQuery(&pageInfo)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	list, total, err := dataPermissionService.GetDataPermissionList(context.Background(), pageInfo)
	if err != nil {
		global.GVA_LOG.Error("获取失败!", zap.Error(err))
		response.FailWithMessage("获取失败:"+err.Error(), c)
		return
	}
	response.OkWithDetailed(response.PageResult{
		List:     list,
		Total:    total,
		Page:     pageInfo.Page,
		PageSize: pageInfo.PageSize,
	}, "获取成功", c)
}
