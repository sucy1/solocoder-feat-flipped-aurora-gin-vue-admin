package request

import (
	"github.com/flipped-aurora/gin-vue-admin/server/model/common/request"
	"github.com/flipped-aurora/gin-vue-admin/server/model/system"
)

type SysDataPermissionCreate struct {
	AuthorityID      uint   `json:"authorityId" binding:"required"`
	Level            string `json:"level" binding:"required,oneof=dept user role custom"`
	CustomConditions string `json:"customConditions"`
}

type SysDataPermissionUpdate struct {
	ID               uint   `json:"ID" binding:"required"`
	Level            string `json:"level" binding:"required,oneof=dept user role custom"`
	CustomConditions string `json:"customConditions"`
}

type SysDataPermissionSearch struct {
	system.SysDataPermission
	request.PageInfo
}
