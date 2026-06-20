package request

import "github.com/flipped-aurora/gin-vue-admin/server/model/common/request"

type SysFieldPermissionCreate struct {
	Path        string `json:"path" binding:"required"`
	Field       string `json:"field" binding:"required"`
	Action      string `json:"action" binding:"required,oneof=read write"`
	AuthorityID uint   `json:"authorityId" binding:"required"`
}

type SysFieldPermissionUpdate struct {
	ID          uint   `json:"ID" binding:"required"`
	Path        string `json:"path" binding:"required"`
	Field       string `json:"field" binding:"required"`
	Action      string `json:"action" binding:"required,oneof=read write"`
	AuthorityID uint   `json:"authorityId" binding:"required"`
}

type SysFieldPermissionSearch struct {
	request.PageInfo
	Path        string `json:"path" form:"path"`
	Field       string `json:"field" form:"field"`
	AuthorityID uint   `json:"authorityId" form:"authorityId"`
}
