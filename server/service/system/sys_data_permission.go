package system

import (
	"context"
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

func (d *DataPermissionService) CreateDataPermission(ctx context.Context, dp system.SysDataPermission) error {
	return global.GVA_DB.Create(&dp).Error
}

func (d *DataPermissionService) UpdateDataPermission(ctx context.Context, dp system.SysDataPermission) error {
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
		return db.Where("user_id IN (SELECT id FROM sys_users WHERE authority_id IN (?))", authorityIds).Or("user_id = ?", userId)
	case "role":
		return db.Where("user_id IN (SELECT id FROM sys_users WHERE authority_id IN (?))", authorityIds)
	case "custom":
		if strings.TrimSpace(customSQL) != "" {
			return db.Where(customSQL)
		}
		return db
	default:
		return db
	}
}
