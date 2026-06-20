package system

import "github.com/flipped-aurora/gin-vue-admin/server/global"

type SysDataPermission struct {
	global.GVA_MODEL
	AuthorityID      uint         `json:"authorityId" gorm:"uniqueIndex;comment:角色ID"`
	Level            string       `json:"level" gorm:"column:level;comment:权限级别(dept/user/role/custom);size:20"`
	CustomConditions string       `json:"customConditions" gorm:"column:custom_conditions;type:text;comment:自定义条件(JSON)"`
	CustomSQL        string       `json:"customSQL" gorm:"column:custom_sql;type:text;comment:自定义SQL(已废弃,请用customConditions)"`
	Authority        SysAuthority `json:"authority" gorm:"foreignKey:AuthorityID;references:AuthorityId"`
}

type ConditionItem struct {
	Field string `json:"field"`
	Op    string `json:"op"`
	Value string `json:"value"`
}
