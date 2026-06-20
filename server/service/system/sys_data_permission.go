package system

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
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

var allowedOps = map[string]bool{
	"=": true, "!=": true, ">": true, "<": true, ">=": true, "<=": true,
	"like": true, "in": true, "not in": true,
	"between": true, "is null": true, "is not null": true,
}

var fieldNamePattern = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`)

func validateCondition(cond system.ConditionItem) error {
	if !fieldNamePattern.MatchString(cond.Field) {
		return fmt.Errorf("invalid field name: %s", cond.Field)
	}
	if !allowedOps[cond.Op] {
		return fmt.Errorf("unsupported operator: %s", cond.Op)
	}
	if cond.Op == "between" {
		parts := strings.Split(cond.Value, ",")
		if len(parts) != 2 {
			return errors.New("between value must be two comma-separated values")
		}
	}
	return nil
}

func validateCustomConditions(conditionsJSON string) error {
	trimmed := strings.TrimSpace(conditionsJSON)
	if trimmed == "" {
		return nil
	}
	var conditions []system.ConditionItem
	if err := json.Unmarshal([]byte(trimmed), &conditions); err != nil {
		return fmt.Errorf("invalid JSON: %v", err)
	}
	for _, cond := range conditions {
		if err := validateCondition(cond); err != nil {
			return err
		}
	}
	return nil
}

func buildConditionQuery(db *gorm.DB, conditionsJSON string) *gorm.DB {
	trimmed := strings.TrimSpace(conditionsJSON)
	if trimmed == "" {
		return db
	}
	var conditions []system.ConditionItem
	if err := json.Unmarshal([]byte(trimmed), &conditions); err != nil {
		return db.Where("1 = 0")
	}
	for _, cond := range conditions {
		if err := validateCondition(cond); err != nil {
			return db.Where("1 = 0")
		}
		quotedField := fmt.Sprintf("`%s`", cond.Field)
		switch cond.Op {
		case "is null":
			db = db.Where(fmt.Sprintf("%s IS NULL", quotedField))
		case "is not null":
			db = db.Where(fmt.Sprintf("%s IS NOT NULL", quotedField))
		case "in", "not in":
			parts := strings.Split(cond.Value, ",")
			placeholders := make([]string, len(parts))
			args := make([]interface{}, len(parts))
			for i, p := range parts {
				placeholders[i] = "?"
				args[i] = strings.TrimSpace(p)
			}
			db = db.Where(fmt.Sprintf("%s %s (%s)", quotedField, strings.ToUpper(cond.Op), strings.Join(placeholders, ",")), args...)
		case "between":
			parts := strings.Split(cond.Value, ",")
			if len(parts) == 2 {
				db = db.Where(fmt.Sprintf("%s BETWEEN ? AND ?", quotedField), strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1]))
			}
		case "like":
			db = db.Where(fmt.Sprintf("%s LIKE ?", quotedField), cond.Value)
		default:
			db = db.Where(fmt.Sprintf("%s %s ?", quotedField, cond.Op), cond.Value)
		}
	}
	return db
}

func (d *DataPermissionService) CreateDataPermission(ctx context.Context, dp system.SysDataPermission) error {
	if err := validateCustomConditions(dp.CustomConditions); err != nil {
		return err
	}
	return global.GVA_DB.Create(&dp).Error
}

func (d *DataPermissionService) UpdateDataPermission(ctx context.Context, dp system.SysDataPermission) error {
	if err := validateCustomConditions(dp.CustomConditions); err != nil {
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
	var customConditions string
	for _, perm := range permissions {
		score, ok := levelStrictness[perm.Level]
		if !ok {
			continue
		}
		if score > strictestScore {
			strictestScore = score
			strictestLevel = perm.Level
			customConditions = perm.CustomConditions
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
		return buildConditionQuery(db, customConditions)
	default:
		return db
	}
}
