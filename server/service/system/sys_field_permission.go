package system

import (
	"context"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/system"
	systemReq "github.com/flipped-aurora/gin-vue-admin/server/model/system/request"
)

type FieldPermissionService struct{}

var FieldPermissionServiceApp = new(FieldPermissionService)

func (s *FieldPermissionService) CreateFieldPermission(ctx context.Context, fp system.SysFieldPermission) error {
	return global.GVA_DB.Create(&fp).Error
}

func (s *FieldPermissionService) UpdateFieldPermission(ctx context.Context, fp system.SysFieldPermission) error {
	return global.GVA_DB.Model(&system.SysFieldPermission{}).Where("id = ?", fp.ID).Updates(&fp).Error
}

func (s *FieldPermissionService) DeleteFieldPermission(ctx context.Context, id uint) error {
	return global.GVA_DB.Delete(&system.SysFieldPermission{}, id).Error
}

func (s *FieldPermissionService) GetFieldPermission(ctx context.Context, id uint) (fp system.SysFieldPermission, err error) {
	err = global.GVA_DB.Where("id = ?", id).First(&fp).Error
	return
}

func (s *FieldPermissionService) GetFieldPermissionList(ctx context.Context, info systemReq.SysFieldPermissionSearch) (list []system.SysFieldPermission, total int64, err error) {
	limit := info.PageSize
	offset := info.PageSize * (info.Page - 1)
	db := global.GVA_DB.Model(&system.SysFieldPermission{})
	if info.Path != "" {
		db = db.Where("path = ?", info.Path)
	}
	if info.Field != "" {
		db = db.Where("field = ?", info.Field)
	}
	if info.AuthorityID != 0 {
		db = db.Where("authority_id = ?", info.AuthorityID)
	}
	err = db.Count(&total).Error
	if err != nil {
		return
	}
	if limit != 0 {
		db = db.Limit(limit).Offset(offset)
	}
	err = db.Find(&list).Error
	return
}

func (s *FieldPermissionService) GetFieldPermissionsForUser(ctx context.Context, path string, authorityIds []uint) (readFields []string, writeFields []string, err error) {
	if len(authorityIds) == 0 {
		return nil, nil, nil
	}
	var count int64
	err = global.GVA_DB.Model(&system.SysFieldPermission{}).Where("path = ?", path).Count(&count).Error
	if err != nil {
		return nil, nil, err
	}
	if count == 0 {
		return nil, nil, nil
	}
	var permissions []system.SysFieldPermission
	err = global.GVA_DB.Where("path = ? AND authority_id IN ?", path, authorityIds).Find(&permissions).Error
	if err != nil {
		return nil, nil, err
	}
	if len(permissions) == 0 {
		return []string{}, []string{}, nil
	}
	readMap := make(map[string]struct{})
	writeMap := make(map[string]struct{})
	for _, perm := range permissions {
		switch perm.Action {
		case "read":
			readMap[perm.Field] = struct{}{}
		case "write":
			writeMap[perm.Field] = struct{}{}
			readMap[perm.Field] = struct{}{}
		}
	}
	for field := range readMap {
		readFields = append(readFields, field)
	}
	for field := range writeMap {
		writeFields = append(writeFields, field)
	}
	return
}

func (s *FieldPermissionService) FilterResponseFields(ctx context.Context, path string, authorityIds []uint, data map[string]interface{}) (map[string]interface{}, error) {
	readFields, _, err := s.GetFieldPermissionsForUser(ctx, path, authorityIds)
	if err != nil {
		return nil, err
	}
	if readFields == nil {
		return data, nil
	}
	readSet := make(map[string]struct{}, len(readFields))
	for _, f := range readFields {
		readSet[f] = struct{}{}
	}
	filtered := make(map[string]interface{}, len(data))
	for k, v := range data {
		if _, ok := readSet[k]; ok {
			filtered[k] = v
		}
	}
	return filtered, nil
}
