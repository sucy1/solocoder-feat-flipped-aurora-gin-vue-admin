package system

import (
	"context"
	"errors"
	"regexp"
	"strings"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/system"
	systemReq "github.com/flipped-aurora/gin-vue-admin/server/model/system/request"
	"gorm.io/gorm"
)

type DataPermissionService struct{}

var DataPermissionServiceApp = new(DataPermissionService)

var levelStrictness = map[string]int{
	"user":   4,
	"dept":   3,
	"role":   2,
	"custom": 1,
}

var dangerousSQLPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)\bunion\b`),
	regexp.MustCompile(`(?i)\bdrop\b`),
	regexp.MustCompile(`(?i)\bdelete\b`),
	regexp.MustCompile(`(?i)\bupdate\b`),
	regexp.MustCompile(`(?i)\binsert\b`),
	regexp.MustCompile(`(?i)\balter\b`),
	regexp.MustCompile(`(?i)\btruncate\b`),
	regexp.MustCompile(`(?i)\binformation_schema\b`),
	regexp.MustCompile(`(?i)\bsysdatabases\b`),
	regexp.MustCompile(`(?i)\bmssql\b`),
	regexp.MustCompile(`(?i);`),
	regexp.MustCompile(`--`),
	regexp.MustCompile(`/\*`),
	regexp.MustCompile(`\*/`),
	regexp.MustCompile(`xp_`),
	regexp.MustCompile(`(?i)\bexec\b`),
	regexp.MustCompile(`(?i)\bexecute\b`),
	regexp.MustCompile(`(?i)\bcreate\b`),
}

var allowedSQLFunctions = map[string]bool{
	"where": true, "and": true, "or": true, "not": true,
	"in": true, "like": true, "between": true, "is": true, "null": true,
	"select": true, "from": true, "exists": true, "case": true, "when": true,
	"then": true, "end": true, "as": true, "on": true,
	"=": true, "!=": true, ">": true, "<": true, ">=": true, "<=": true,
	"<>": true,
}

func validateCustomSQL(sql string) error {
	trimmed := strings.TrimSpace(sql)
	if trimmed == "" {
		return nil
	}
	for _, pattern := range dangerousSQLPatterns {
		if pattern.MatchString(trimmed) {
			return errors.New("custom SQL contains dangerous patterns")
		}
	}
	parenDepth := 0
	for _, ch := range trimmed {
		switch ch {
		case '(':
			parenDepth++
		case ')':
			parenDepth--
			if parenDepth < 0 {
				return errors.New("custom SQL has unbalanced parentheses")
			}
		}
	}
	if parenDepth != 0 {
		return errors.New("custom SQL has unbalanced parentheses")
	}
	return nil
}

func (d *DataPermissionService) CreateDataPermission(ctx context.Context, dp system.SysDataPermission) error {
	if err := validateCustomSQL(dp.CustomSQL); err != nil {
		return err
	}
	return global.GVA_DB.Create(&dp).Error
}

func (d *DataPermissionService) UpdateDataPermission(ctx context.Context, dp system.SysDataPermission) error {
	if err := validateCustomSQL(dp.CustomSQL); err != nil {
		return err
	}
	return global.GVA_DB.Model(&system.SysDataPermission{}).Where("id = ?", dp.ID).Updates(&dp).Error
}

func (d *DataPermissionService) DeleteDataPermission(ctx context.Context, id uint) error {
	return global.GVA_DB.Delete(&system.SysDataPermission{}, id).Error
}

func (d *DataPermissionService) GetDataPermission(ctx context.Context, id uint) (dp system.SysDataPermission, err error) {
	err = global.GVA_DB.Where("id = ?", id).Preload("Authority").First(&dp).Error
	return
}

func (d *DataPermissionService) GetDataPermissionList(ctx context.Context, info systemReq.SysDataPermissionSearch) (list []system.SysDataPermission, total int64, err error) {
	limit := info.PageSize
	offset := info.PageSize * (info.Page - 1)
	db := global.GVA_DB.Model(&system.SysDataPermission{})
	err = db.Count(&total).Error
	if err != nil {
		return
	}
	if limit != 0 {
		db = db.Limit(limit).Offset(offset)
	}
	err = db.Preload("Authority").Find(&list).Error
	return
}

func (d *DataPermissionService) GetDataPermissionByAuthorityId(ctx context.Context, authorityId uint) (dp system.SysDataPermission, err error) {
	err = global.GVA_DB.Where("authority_id = ?", authorityId).Preload("Authority").First(&dp).Error
	return
}

func (d *DataPermissionService) ApplyDataPermission(db *gorm.DB, userId uint, authorityIds []uint) *gorm.DB {
	if len(authorityIds) == 0 {
		return db
	}
	var permissions []system.SysDataPermission
	err := global.GVA_DB.Where("authority_id IN ?", authorityIds).Find(&permissions).Error
	if err != nil || len(permissions) == 0 {
		return db
	}
	strictestLevel := ""
	strictestScore := 0
	var customSQL string
	for _, perm := range permissions {
		score, ok := levelStrictness[perm.Level]
		if !ok {
			continue
		}
		if score > strictestScore {
			strictestScore = score
			strictestLevel = perm.Level
			customSQL = perm.CustomSQL
		}
	}
	switch strictestLevel {
	case "user":
		return db.Where("user_id = ?", userId)
	case "dept":
		return db.Where("user_id IN (SELECT id FROM sys_users WHERE authority_id IN (?)) OR user_id = ?", authorityIds, userId)
	case "role":
		return db.Where("user_id IN (SELECT id FROM sys_users WHERE authority_id IN (?))", authorityIds)
	case "custom":
		trimmed := strings.TrimSpace(customSQL)
		if trimmed == "" {
			return db
		}
		if err := validateCustomSQL(trimmed); err != nil {
			return db.Where("1 = 0")
		}
		return db.Where("(" + trimmed + ")")
	default:
		return db
	}
}
