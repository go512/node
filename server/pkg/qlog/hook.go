package qlog

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"runtime"
	"strings"
)

type Hook interface {
	// Levels 返回 Hook 生效的日志级别（如 Debug/Info/Error）
	Levels() []logrus.Level
	// Fire 日志输出时触发的逻辑（核心：修改 entry 或增强日志）
	Fire(*logrus.Entry) error
}

// CallerSkipHook 解决 logrus.SetReportCaller 跳过层数不准的问题
// 支持自定义 skip 层数，精准获取业务代码的文件/行号
type CallerSkipHook struct {
	Skip int //自定义跳过的调用栈层数
}

// Levels 该 Hook 对所有级别生效
func (h *CallerSkipHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

// Fire 核心逻辑重写 entry.Caller
func (h *CallerSkipHook) Fire(entry *logrus.Entry) error {
	//跳过指定层数，获得业务代码的文件/行号
	//skip = h.skip + logrus 内部调用层数（固定+2）
	skip := h.Skip + 2
	pc, file, line, ok := runtime.Caller(skip)
	if !ok {
		return nil
	}

	// 简化文件名（只保留最后一段，如 main.go 而非 /xxx/yyy/main.go）
	//file = filepath.Base(file) 暂时不简化
	funcName := runtime.FuncForPC(pc).Name()
	funcName = funcName[strings.LastIndex(funcName, "/")+1:]

	// 重写 entry.Caller （覆盖 logrus 原生的 Caller）
	entry.Caller = &runtime.Frame{
		PC:       pc,
		File:     file,
		Line:     line,
		Function: funcName,
	}

	// 可选：将 caller 信息注入日志字段（方便 JSON 格式解析）
	entry.Data["caller"] = fmt.Sprintf("%s:%d", file, line)
	entry.Data["file"] = file
	entry.Data["line"] = line
	entry.Data["func"] = funcName
	return nil
}

// 自定义BizMetaHook 用于注入业务元信息
type BizMetaHook struct {
	ServiceName string
	Env         string //环境 dev/test/prd
	Version     string //版本号
}

func (h *BizMetaHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

func (h *BizMetaHook) Fire(entry *logrus.Entry) error {
	entry.Data["service"] = h.ServiceName
	entry.Data["env"] = h.Env
	entry.Data["version"] = h.Version
	return nil
}
