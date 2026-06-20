package system

import (
	"context"
	"strconv"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/common/response"
	"github.com/flipped-aurora/gin-vue-admin/server/model/system"
	systemReq "github.com/flipped-aurora/gin-vue-admin/server/model/system/request"
	"github.com/flipped-aurora/gin-vue-admin/server/utils"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type FieldPermissionApi struct{}

func (f *FieldPermissionApi) CreateFieldPermission(c *gin.Context) {
	var fp systemReq.SysFieldPermissionCreate
	err := c.ShouldBindJSON(&fp)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	fieldPermission := system.SysFieldPermission{
		Path:        fp.Path,
		Field:       fp.Field,
		Action:      fp.Action,
		AuthorityID: fp.AuthorityID,
	}
	err = fieldPermissionService.CreateFieldPermission(context.Background(), fieldPermission)
	if err != nil {
		global.GVA_LOG.Error("创建失败!", zap.Error(err))
		response.FailWithMessage("创建失败:"+err.Error(), c)
		return
	}
	response.OkWithMessage("创建成功", c)
}

func (f *FieldPermissionApi) UpdateFieldPermission(c *gin.Context) {
	var fp systemReq.SysFieldPermissionUpdate
	err := c.ShouldBindJSON(&fp)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	fieldPermission := system.SysFieldPermission{
		Path:        fp.Path,
		Field:       fp.Field,
		Action:      fp.Action,
		AuthorityID: fp.AuthorityID,
	}
	fieldPermission.ID = fp.ID
	err = fieldPermissionService.UpdateFieldPermission(context.Background(), fieldPermission)
	if err != nil {
		global.GVA_LOG.Error("更新失败!", zap.Error(err))
		response.FailWithMessage("更新失败:"+err.Error(), c)
		return
	}
	response.OkWithMessage("更新成功", c)
}

func (f *FieldPermissionApi) DeleteFieldPermission(c *gin.Context) {
	idStr := c.Query("ID")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		response.FailWithMessage("无效的ID参数", c)
		return
	}
	err = fieldPermissionService.DeleteFieldPermission(context.Background(), uint(id))
	if err != nil {
		global.GVA_LOG.Error("删除失败!", zap.Error(err))
		response.FailWithMessage("删除失败:"+err.Error(), c)
		return
	}
	response.OkWithMessage("删除成功", c)
}

func (f *FieldPermissionApi) FindFieldPermission(c *gin.Context) {
	idStr := c.Query("ID")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		response.FailWithMessage("无效的ID参数", c)
		return
	}
	reFp, err := fieldPermissionService.GetFieldPermission(context.Background(), uint(id))
	if err != nil {
		global.GVA_LOG.Error("查询失败!", zap.Error(err))
		response.FailWithMessage("查询失败:"+err.Error(), c)
		return
	}
	response.OkWithDetailed(reFp, "查询成功", c)
}

func (f *FieldPermissionApi) GetFieldPermissionList(c *gin.Context) {
	var pageInfo systemReq.SysFieldPermissionSearch
	err := c.ShouldBindQuery(&pageInfo)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	list, total, err := fieldPermissionService.GetFieldPermissionList(context.Background(), pageInfo)
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

func (f *FieldPermissionApi) GetFieldPermissionsForPath(c *gin.Context) {
	path := c.Query("path")
	if path == "" {
		response.FailWithMessage("path参数不能为空", c)
		return
	}
	authorityId := utils.GetUserAuthorityId(c)
	authorityIds := []uint{authorityId}
	if raw, exists := c.Get("data_permission_authority_ids"); exists {
		if ids, ok := raw.([]uint); ok {
			authorityIds = ids
		}
	}
	readFields, writeFields, err := fieldPermissionService.GetFieldPermissionsForUser(context.Background(), path, authorityIds)
	if err != nil {
		global.GVA_LOG.Error("查询字段权限失败!", zap.Error(err))
		response.FailWithMessage("查询字段权限失败:"+err.Error(), c)
		return
	}
	if readFields == nil && writeFields == nil {
		response.OkWithDetailed(gin.H{"allFields": true}, "获取成功", c)
		return
	}
	response.OkWithDetailed(gin.H{
		"read":  readFields,
		"write": writeFields,
	}, "获取成功", c)
}
