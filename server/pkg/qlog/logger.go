package qlog

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
)

const (
	JsonFormat = "json"
	TextFormat = "text"

	errorKey = "error"
)

type Option struct {
	Level             string `json:"level"`                // 日志级别: debug/info/warn/error/fatal/panic
	Output            string `json:"output"`               // 日志输出位置: stdout/stderr/file
	OutputFilePath    string `json:"output_file_path"`     // 日志输出文件路径
	OutputFileMaxSize int64  `json:"output_file_max_size"` // 日志输出文件最大大小 (MB)
	Formatter         string `json:"formatter"`            // 日志格式: text/json
	CallerSkip        int    `json:"caller_skip"`          // 打印具体文件的跳过层数
	EnableCaller      bool   `json:"enable_caller"`        // 是否显示调用者信息
	TimestampFormat   string `json:"timestamp_format"`     // 时间戳格式
	MaxBackups        int    `json:"max_backups"`          // 最大备份文件数
	MaxAge            int    `json:"max_age"`              // 日志文件最大保存天数
	Compress          bool   `json:"compress"`             // 是否压缩旧日志文件
}

var defaultOption = &Option{
	Level:             "debug",
	Output:            "stdout",
	Formatter:         "text",
	CallerSkip:        6,
	EnableCaller:      true,
	TimestampFormat:   "2006-01-02 15:04:05",
	OutputFileMaxSize: 100, // 100MB
	MaxBackups:        3,
	MaxAge:            30, // 30 days
	Compress:          true,
}

type Logger interface {
	SetOutput(output io.Writer)                                  // 设置输出
	GetOutput() io.Writer                                        // 获取输出
	SetLevel(level logrus.Level)                                 // 设置log等级
	GetLevel() logrus.Level                                      // 获取log等级
	Log(level logrus.Level, args ...interface{})                 // 记录对应级别的日志
	Logf(level logrus.Level, format string, args ...interface{}) // 记录对应级别的日志
	Trace(args ...interface{})                                   // 记录 TraceLevel 级别的日志
	Tracef(format string, args ...interface{})                   // 格式化并记录 TraceLevel 级别的日志
	Debug(args ...interface{})                                   // 记录 DebugLevel 级别的日志
	Debugf(format string, args ...interface{})                   // 格式化并记录 DebugLevel 级别的日志
	Info(args ...interface{})                                    // 记录 InfoLevel 级别的日志
	Infof(format string, args ...interface{})                    // 格式化并记录 InfoLevel 级别的日志
	Print(args ...interface{})                                   // 记录 InfoLevel 级别的日志[gorm logger扩展]
	Printf(format string, args ...interface{})                   // 格式化并记录 InfoLevel 级别的日志[gorm logger扩展]
	Warn(args ...interface{})                                    // 记录 WarnLevel 级别的日志
	Warnf(format string, args ...interface{})                    // 格式化并记录 WarnLevel 级别的日志
	Error(args ...interface{})                                   // 记录 ErrorLevel 级别的日志
	Errorf(format string, args ...interface{})                   // 格式化并记录 ErrorLevel 级别的日志
	Panicf(format string, args ...interface{})                   // 格式化并记录 PanicLevel 级别的日志
	WithField(key string, value interface{}) Logger              // 为日志添加一个上下文数据
	WithFields(fields Fields) Logger                             // 为日志添加多个上下文数据
	WithError(err error) Logger                                  // 为日志添加标准错误上下文数据
}

type Fields map[string]interface{}

type logentry struct {
	entry *logrus.Entry
}

// New 创建一个使用默认配置的 Logger
func New() Logger {
	return NewWithOption(defaultOption)
}

func NewWithOption(opt *Option) Logger {
	if opt == nil {
		opt = defaultOption
	}

	// 创建 logrus.Logger 对象
	logger := logrus.New()

	// 设置日志级别
	level, err := logrus.ParseLevel(opt.Level)
	if err != nil {
		level = logrus.InfoLevel
		logger.WithField("invalid_level", opt.Level).Warn("无效的日志级别，使用默认 info 级别")
	}
	logger.SetLevel(level)

	// 设置日志输出
	output := setupOutput(opt, logger)
	log.SetOutput(output)

	// 设置日志格式
	formatter := setupFormatter(opt)
	logger.SetFormatter(formatter)

	// 设置调用者信息和钩子
	setupHooks(logger, opt)

	return &logentry{
		entry: logrus.NewEntry(logger),
	}
}

// setupFormatter 设置日志格式
func setupFormatter(option *Option) logrus.Formatter {
	if option.Formatter == "json" {
		return &logrus.JSONFormatter{
			TimestampFormat: option.TimestampFormat,
			PrettyPrint:     false,
		}
	} else {
		return &logrus.TextFormatter{
			TimestampFormat: option.TimestampFormat,
			FullTimestamp:   true,
			ForceColors:     true,
		}
	}
}

// setupHooks 设置钩子
func setupHooks(logger *logrus.Logger, option *Option) {
	// 设置调用者信息
	if option.EnableCaller {
		logger.AddHook(&CallerSkipHook{
			Skip: option.CallerSkip,
		})
		// 关闭logrus的调用者信息，使用钩子控制
		logger.SetReportCaller(option.EnableCaller)
	}
}

// setupOutput 设置日志输出
func setupOutput(option *Option, logger *logrus.Logger) io.Writer {
	var output io.Writer = os.Stdout

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
			return os.Stdout
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
			file, err := os.OpenFile(option.OutputFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
			if err != nil {
				logger.WithError(err).Error("日志文件打开失败，降级到控制台输出")
				return os.Stdout
			}
			output = file
		}
	default:
		output = os.Stdout
	}

	return output
}

func (l *logentry) SetLevel(level logrus.Level) {
	l.entry.Logger.SetLevel(level)
}

// 获取日志等级
func (l *logentry) GetLevel() logrus.Level {
	return l.entry.Logger.GetLevel()
}

// 设置输出
func (l *logentry) SetOutput(output io.Writer) {
	l.entry.Logger.SetOutput(output)
}

// 获取日志等级
func (l *logentry) GetOutput() io.Writer {
	return l.entry.Logger.Out
}

func (l *logentry) Log(level logrus.Level, args ...interface{}) {
	l.entry.Log(level, args...)
}

func (l *logentry) Logf(level logrus.Level, format string, args ...interface{}) {
	l.entry.Logf(level, format, args...)
}

// 记录一条 LevelDebug 级别的日志
func (l *logentry) Trace(args ...interface{}) {
	l.entry.Log(logrus.TraceLevel, args...)
}

// 格式化并记录一条 LevelDebug 级别的日志
func (l *logentry) Tracef(format string, args ...interface{}) {
	l.entry.Logf(logrus.TraceLevel, format, args...)
}

func (l *logentry) Debug(args ...interface{}) {
	l.entry.Log(logrus.DebugLevel, args...)
}

func (l *logentry) Debugf(format string, args ...interface{}) {
	l.entry.Logf(logrus.DebugLevel, format, args...)
}

func (l *logentry) Info(args ...interface{}) {
	l.entry.Log(logrus.InfoLevel, args...)
}

func (l *logentry) Infof(format string, args ...interface{}) {
	l.entry.Logf(logrus.InfoLevel, format, args...)
}

func (l *logentry) Print(args ...interface{}) {
	l.entry.Log(logrus.InfoLevel, args...)
}

func (l *logentry) Printf(format string, args ...interface{}) {
	l.entry.Logf(logrus.InfoLevel, format, args...)
}

func (l *logentry) Warn(args ...interface{}) {
	l.entry.Log(logrus.WarnLevel, args...)
}

func (l *logentry) Warnf(format string, args ...interface{}) {
	l.entry.Logf(logrus.WarnLevel, format, args...)
}

func (l *logentry) Error(args ...interface{}) {
	l.entry.Log(logrus.ErrorLevel, args...)
}

func (l *logentry) Errorf(format string, args ...interface{}) {
	l.entry.Logf(logrus.ErrorLevel, format, args...)
}

func (l *logentry) Panicf(format string, args ...interface{}) {
	l.entry.Logf(logrus.PanicLevel, format, args...)
	panic(fmt.Sprintf(format, args...))
}

// 为当前日志附加一组上下文数据
func (l *logentry) WithFields(fields Fields) Logger {
	if err, ok := fields[errorKey].(interface {
		Stack() []string
	}); ok {
		fields["err.stack"] = strings.Join(err.Stack(), ";")
	}

	return &logentry{entry: l.entry.WithFields(logrus.Fields(fields))}
}

func (l *logentry) WithField(key string, value interface{}) Logger {
	return l.WithFields(Fields{key: value})
}

// 为当前日志附加一个错误
func (l *logentry) WithError(err error) Logger {
	return l.WithFields(Fields{errorKey: err})
}
