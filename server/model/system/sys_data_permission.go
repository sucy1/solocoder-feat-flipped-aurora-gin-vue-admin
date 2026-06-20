package system

import "github.com/flipped-aurora/gin-vue-admin/server/global"

type SysDataPermission struct {
	global.GVA_MODEL
	AuthorityID uint         `json:"authorityId" gorm:"uniqueIndex;comment:角色ID"`
	Level       string       `json:"level" gorm:"column:level;comment:权限级别(dept/user/role/custom);size:20"`
	CustomSQL   string       `json:"customSQL" gorm:"column:custom_sql;type:text;comment:自定义SQL WHERE条件"`
	Authority   SysAuthority `json:"authority" gorm:"foreignKey:AuthorityID;references:AuthorityId"`
}
