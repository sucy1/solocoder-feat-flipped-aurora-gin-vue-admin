package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/system"
	"github.com/flipped-aurora/gin-vue-admin/server/utils"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

var operLogChan = make(chan system.SysOperLog, 1024)

func init() {
	go operLogWorker()
}

func operLogWorker() {
	for log := range operLogChan {
		if err := global.GVA_DB.Create(&log).Error; err != nil {
			global.GVA_LOG.Error("create oper log error:", zap.Error(err))
		}
	}
}

func maskSensitiveFields(body string) string {
	maskFields := global.GVA_CONFIG.System.OperLogMaskFields
	if len(maskFields) == 0 {
		return body
	}
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(body), &data); err != nil {
		return body
	}
	changed := false
	for _, field := range maskFields {
		if _, ok := data[field]; ok {
			data[field] = "***"
			changed = true
		}
	}
	if !changed {
		return body
	}
	masked, err := json.Marshal(&data)
	if err != nil {
		return body
	}
	return string(masked)
}

func shouldSkipPath(path string) bool {
	skipPaths := global.GVA_CONFIG.System.OperLogSkipPaths
	for _, p := range skipPaths {
		if strings.HasPrefix(path, p) {
			return true
		}
	}
	return false
}

const maxBodyLen = 2048

func truncateBody(s string) string {
	if len(s) > maxBodyLen {
		return s[:maxBodyLen]
	}
	return s
}

func OperLog() gin.HandlerFunc {
	return func(c *gin.Context) {
		if _, exists := c.Get("skip_oper_log"); exists {
			c.Next()
			return
		}

		path := c.Request.URL.Path
		if shouldSkipPath(path) {
			c.Next()
			return
		}

		var body []byte
		var userId int
		if c.Request.Method != http.MethodGet {
			var err error
			body, err = io.ReadAll(c.Request.Body)
			if err != nil {
				global.GVA_LOG.Error("read body from request error:", zap.Error(err))
			} else {
				c.Request.Body = io.NopCloser(bytes.NewBuffer(body))
			}
			if len(body) == 0 {
				c.Next()
				return
			}
		} else {
			rawQuery := c.Request.URL.RawQuery
			if rawQuery == "" {
				c.Next()
				return
			}
			query, _ := url.QueryUnescape(rawQuery)
			split := strings.Split(query, "&")
			m := make(map[string]string)
			for _, v := range split {
				kv := strings.Split(v, "=")
				if len(kv) == 2 {
					m[kv[0]] = kv[1]
				}
			}
			if len(m) == 0 {
				c.Next()
				return
			}
			body, _ = json.Marshal(&m)
		}
		claims, _ := utils.GetClaims(c)
		if claims != nil && claims.BaseClaims.ID != 0 {
			userId = int(claims.BaseClaims.ID)
		} else {
			id, err := strconv.Atoi(c.Request.Header.Get("x-user-id"))
			if err != nil {
				userId = 0
			}
			userId = id
		}

		writer := responseBodyWriter{
			ResponseWriter: c.Writer,
			body:           &bytes.Buffer{},
		}
		c.Writer = writer
		now := time.Now()

		c.Next()

		latency := time.Since(now)

		record := system.SysOperLog{
			Ip:           c.ClientIP(),
			Method:       c.Request.Method,
			Path:         path,
			Status:       c.Writer.Status(),
			Latency:      latency,
			Agent:        c.Request.UserAgent(),
			ErrorMessage: c.Errors.ByType(gin.ErrorTypePrivate).String(),
			UserID:       userId,
		}

		if strings.Contains(c.GetHeader("Content-Type"), "multipart/form-data") {
			record.Body = "[文件]"
		} else {
			record.Body = truncateBody(maskSensitiveFields(string(body)))
		}
		record.Resp = truncateBody(writer.body.String())

		select {
		case operLogChan <- record:
		default:
			global.GVA_LOG.Error("oper log channel is full, dropping log")
		}
	}
}
