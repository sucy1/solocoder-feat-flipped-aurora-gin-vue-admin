package system

import (
	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/common/request"
	"github.com/flipped-aurora/gin-vue-admin/server/model/system"
	systemReq "github.com/flipped-aurora/gin-vue-admin/server/model/system/request"
)

type OperLogService struct{}

var OperLogServiceApp = new(OperLogService)

func (operLogService *OperLogService) CreateSysOperLog(sysOperLog system.SysOperLog) error {
	err := global.GVA_DB.Create(&sysOperLog).Error
	return err
}

func (operLogService *OperLogService) DeleteSysOperLogByIds(ids request.IdsReq) error {
	err := global.GVA_DB.Delete(&[]system.SysOperLog{}, "id in (?)", ids.Ids).Error
	return err
}

func (operLogService *OperLogService) DeleteSysOperLog(sysOperLog system.SysOperLog) error {
	err := global.GVA_DB.Delete(&sysOperLog).Error
	return err
}

func (operLogService *OperLogService) GetSysOperLog(id uint) (sysOperLog system.SysOperLog, err error) {
	err = global.GVA_DB.Where("id = ?", id).First(&sysOperLog).Error
	return
}

func (operLogService *OperLogService) GetSysOperLogInfoList(info systemReq.SysOperLogSearch) (list interface{}, total int64, err error) {
	limit := info.PageSize
	offset := info.PageSize * (info.Page - 1)
	db := global.GVA_DB.Model(&system.SysOperLog{})
	var sysOperLogs []system.SysOperLog
	if info.Method != "" {
		db = db.Where("method = ?", info.Method)
	}
	if info.Path != "" {
		db = db.Where("path LIKE ?", "%"+info.Path+"%")
	}
	if info.Status != 0 {
		db = db.Where("status = ?", info.Status)
	}
	err = db.Count(&total).Error
	if err != nil {
		return
	}
	err = db.Order("id desc").Limit(limit).Offset(offset).Preload("User").Find(&sysOperLogs).Error
	return sysOperLogs, total, err
}
