package qlog

import (
	"github.com/sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"
	"io"
	"os"
	"path/filepath"
)

type Option struct {
	Level             string `json:"level"`                // 日志级别: debug/info/warn/error/fatal/panic
	Output            string `json:"output"`               // 日志输出位置: stdout/stderr/file
	OutputFilePath    string `json:"output_file_path"`     // 日志输出文件路径
	OutputFileMaxSize int64  `json:"output_file_max_size"` // 日志输出文件最大大小 (MB)
	Formatter         string `json:"formatter"`            // 日志格式: text/json
	CallerSkip        int    `json:"caller_skip"`          // 打印具体文件的跳过层数
	EnableHTMLEscape  bool   `json:"enable_html_escape"`   // 是否开启html转义
	EnableCaller      bool   `json:"enable_caller"`        // 是否显示调用者信息
	TimestampFormat   string `json:"timestamp_format"`     // 时间戳格式
	MaxBackups        int    `json:"max_backups"`          // 最大备份文件数
	MaxAge            int    `json:"max_age"`              // 日志文件最大保存天数
	Compress          bool   `json:"compress"`             // 是否压缩旧日志文件
}

var defaultOption = &Option{
	Level:             "info",
	Output:            "stdout",
	Formatter:         "text",
	CallerSkip:        4,
	EnableHTMLEscape:  false,
	EnableCaller:      true,
	TimestampFormat:   "2006-01-02 15:04:05",
	OutputFileMaxSize: 100, // 100MB
	MaxBackups:        3,
	MaxAge:            30, // 30 days
	Compress:          true,
}

// Logger 扩展 logrus 接口，提供额外的便捷方法
type Logger interface {
	logrus.Ext1FieldLogger
	SetOutput(output io.Writer)
	SetFormatter(formatter logrus.Formatter)
	SetLevel(level logrus.Level)
	GetLevel() logrus.Level
	AddHook(hook logrus.Hook)
	IsLevelEnabled(level logrus.Level) bool
	WithField(key string, value interface{}) *logrus.Entry
	WithFields(fields logrus.Fields) *logrus.Entry
	WithError(err error) *logrus.Entry
	Debugf(format string, args ...interface{})
	Infof(format string, args ...interface{})
	Printf(format string, args ...interface{})
	Warnf(format string, args ...interface{})
	Warningf(format string, args ...interface{})
	Errorf(format string, args ...interface{})
	Fatalf(format string, args ...interface{})
	Panicf(format string, args ...interface{})
	Debug(args ...interface{})
	Info(args ...interface{})
	Print(args ...interface{})
	Warn(args ...interface{})
	Warning(args ...interface{})
	Error(args ...interface{})
	Fatal(args ...interface{})
	// WithComponent 添加组件字段
	WithComponent(component string) *logrus.Entry
	// WithRequestID adds a request ID field to the log entry
	WithRequestID(requestID string) *logrus.Entry
	// Close closes the log file if one was created
	Close() error
}

// New 创建一个使用默认配置的 Logger
func New() Logger {
	return NewWithOption(defaultOption)
}

// NewWithOption 根据选项创建 Logger
func NewWithOption(option *Option) Logger {
	if option == nil {
		option = defaultOption
	}

	// 创建 logrus.Logger 对象
	logger := logrus.New()

	// 设置日志级别
	level, err := logrus.ParseLevel(option.Level)
	if err != nil {
		level = logrus.InfoLevel
		logger.WithField("invalid_level", option.Level).Warn("无效的日志级别，使用默认 info 级别")
	}
	logger.SetLevel(level)

	// 设置日志输出
	var output io.Writer = os.Stdout
	var file *os.File

	switch option.Output {
	case "stderr":
		output = os.Stderr
	case "file":
		if option.OutputFilePath == "" {
			option.OutputFilePath = "logs/app.log"
		}

		// 确保日志目录存在
		if err := os.MkdirAll(filepath.Dir(option.OutputFilePath), 0755); err != nil {
			logger.WithError(err).Error("创建日志目录失败，使用控制台输出")
			output = os.Stdout
			break
		}

		// 使用 lumberjack 实现日志轮转
		if option.OutputFileMaxSize > 0 {
			output = &lumberjack.Logger{
				Filename:   option.OutputFilePath,
				MaxSize:    int(option.OutputFileMaxSize), // MB
				MaxBackups: option.MaxBackups,
				MaxAge:     option.MaxAge,
				Compress:   option.Compress,
			}
		} else {
			file, err = os.OpenFile(option.OutputFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
			if err != nil {
				logger.WithError(err).Error("日志文件打开失败，降级到控制台输出")
				output = os.Stdout
			} else {
				output = file
			}
		}
	default:
		output = os.Stdout
	}

	logger.SetOutput(output)

	// 设置日志格式
	if option.Formatter == "json" {
		logger.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: option.TimestampFormat,
			PrettyPrint:     false,
		})
	} else {
		logger.SetFormatter(&logrus.TextFormatter{
			TimestampFormat: option.TimestampFormat,
			FullTimestamp:   true,
			ForceColors:     true,
		})
	}

	// 设置调用者信息
	logger.SetReportCaller(option.EnableCaller)
	if option.EnableCaller {
		logger.AddHook(&CallerSkipHook{
			Skip: option.CallerSkip,
		})
		//关闭logurs的调用者信息
		logger.SetReportCaller(false)
	}

	//注册BizErrorHook
	logger.AddHook(&BizMetaHook{
		ServiceName: "test",
		Env:         "test",
	})

	return &loggerWrapper{
		Logger: logger,
		file:   file,
	}
}

// loggerWrapper 包装 logrus.Logger，提供额外功能
type loggerWrapper struct {
	*logrus.Logger
	file *os.File
}

// WithComponent 添加组件字段
func (l *loggerWrapper) WithComponent(component string) *logrus.Entry {
	return l.WithField("component", component)
}

// WithRequestID 添加请求ID字段
func (l *loggerWrapper) WithRequestID(requestID string) *logrus.Entry {
	return l.WithField("request_id", requestID)
}

// Close 关闭日志文件
func (l *loggerWrapper) Close() error {
	if l.file != nil {
		return l.file.Close()
	}
	return nil
}
