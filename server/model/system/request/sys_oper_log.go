package request

import "github.com/flipped-aurora/gin-vue-admin/server/model/common/request"

type SysOperLogSearch struct {
	request.PageInfo
	Method string `json:"method" form:"method"`
	Path   string `json:"path" form:"path"`
	Status int    `json:"status" form:"status"`
}
