package system

import (
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
)

type SysOperLog struct {
	global.GVA_MODEL
	Ip           string        `json:"ip" gorm:"column:ip;comment:请求ip;size:128"`
	Method       string        `json:"method" gorm:"column:method;comment:请求方法;size:16"`
	Path         string        `json:"path" gorm:"column:path;comment:请求路径;size:255"`
	Status       int           `json:"status" gorm:"column:status;comment:请求状态"`
	Latency      time.Duration `json:"latency" gorm:"column:latency;comment:延迟" swaggertype:"string"`
	Agent        string        `json:"agent" gorm:"type:text;column:agent;comment:代理"`
	Body         string        `json:"body" gorm:"type:text;column:body;comment:请求Body"`
	Resp         string        `json:"resp" gorm:"type:text;column:resp;comment:响应Body"`
	ErrorMessage string        `json:"error_message" gorm:"column:error_message;comment:错误信息;size:255"`
	UserID       int           `json:"user_id" gorm:"column:user_id;comment:用户id"`
	User         SysUser       `json:"user" gorm:"foreignKey:UserID"`
}
