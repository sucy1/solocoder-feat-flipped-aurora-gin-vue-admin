package system

import "github.com/flipped-aurora/gin-vue-admin/server/global"

type SysFieldPermission struct {
	global.GVA_MODEL
	Path        string `json:"path" gorm:"column:path;comment:API路径;size:255;uniqueIndex:idx_path_field_action"`
	Field       string `json:"field" gorm:"column:field;comment:字段名;size:100;uniqueIndex:idx_path_field_action"`
	Action      string `json:"action" gorm:"column:action;comment:操作(read/write);size:10;uniqueIndex:idx_path_field_action"`
	AuthorityID uint   `json:"authorityId" gorm:"column:authority_id;comment:角色ID;uniqueIndex:idx_path_field_action"`
}
