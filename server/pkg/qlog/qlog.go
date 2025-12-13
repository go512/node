package qlog

import (
	"github.com/sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"
	"io"
	"os"
	"path/filepath"
	"sync"
)

// Logger 统一日志接口
type Logger interface {
	// 基础日志方法
	Debug(args ...interface{})
	Debugf(format string, args ...interface{})
	Info(args ...interface{})
	Infof(format string, args ...interface{})
	Warn(args ...interface{})
	Warnf(format string, args ...interface{})
	Error(args ...interface{})
	Errorf(format string, args ...interface{})
	Fatal(args ...interface{})
	Fatalf(format string, args ...interface{})
	Panic(args ...interface{})
	Panicf(format string, args ...interface{})

	// 字段方法
	WithField(key string, value interface{}) Logger
	WithFields(fields map[string]interface{}) Logger
	WithError(err error) Logger

	// 配置方法
	SetLevel(level logrus.Level)
	GetLevel() logrus.Level
	SetOutput(output io.Writer)
	GetOutput() io.Writer
	SetFormatter(formatter logrus.Formatter)

	// 扩展方法
	WithComponent(component string) Logger
	WithRequestID(requestID string) Logger

	// 资源管理
	Close() error
}

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
	ServiceName       string `json:"service_name"`         // 服务名称
	Env               string `json:"env"`                  // 环境
}

var defaultOption = &Option{
	Level:             "info",
	Output:            "stdout",
	Formatter:         "text",
	CallerSkip:        4,
	EnableCaller:      true,
	TimestampFormat:   "2006-01-02 15:04:05",
	OutputFileMaxSize: 100, // 100MB
	MaxBackups:        3,
	MaxAge:            30, // 30 days
	Compress:          true,
	ServiceName:       "unknown",
	Env:               "development",
}

// loggerImpl Logger接口的具体实现
type loggerImpl struct {
	entry *logrus.Entry
	file  *os.File
	mu    sync.RWMutex
}

// 全局日志实例
var (
	GlobalLogger Logger
	once         sync.Once
)

// InitLogger 初始化全局日志实例（单例模式）
func InitLogger(option *Option) {
	once.Do(func() {
		GlobalLogger = NewWithOption(option)
	})
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

	// 深拷贝选项避免修改默认配置
	opt := *option

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
	output, file := setupOutput(&opt, logger)
	logger.SetOutput(output)

	// 设置日志格式
	formatter := setupFormatter(&opt)
	logger.SetFormatter(formatter)

	// 设置调用者信息和钩子
	setupHooks(logger, &opt)

	return &loggerImpl{
		entry: logrus.NewEntry(logger),
		file:  file,
	}
}

// setupOutput 设置日志输出
func setupOutput(option *Option, logger *logrus.Logger) (io.Writer, *os.File) {
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
			return os.Stdout, nil
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
				return os.Stdout, nil
			}
			output = file
		}
	default:
		output = os.Stdout
	}

	return output, file
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
		logger.SetReportCaller(false)
	}

	// 注册业务元数据钩子
	logger.AddHook(&BizMetaHook{
		ServiceName: option.ServiceName,
		Env:         option.Env,
	})
}

// 实现Logger接口的基础方法
func (l *loggerImpl) Debug(args ...interface{}) {
	l.mu.RLock()
	l.entry.Debug(args...)
	l.mu.RUnlock()
}

func (l *loggerImpl) Debugf(format string, args ...interface{}) {
	l.mu.RLock()
	l.entry.Debugf(format, args...)
	l.mu.RUnlock()
}

func (l *loggerImpl) Info(args ...interface{}) {
	l.mu.RLock()
	l.entry.Info(args...)
	l.mu.RUnlock()
}

func (l *loggerImpl) Infof(format string, args ...interface{}) {
	l.mu.RLock()
	l.entry.Infof(format, args...)
	l.mu.RUnlock()
}

func (l *loggerImpl) Warn(args ...interface{}) {
	l.mu.RLock()
	l.entry.Warn(args...)
	l.mu.RUnlock()
}

func (l *loggerImpl) Warnf(format string, args ...interface{}) {
	l.mu.RLock()
	l.entry.Warnf(format, args...)
	l.mu.RUnlock()
}

func (l *loggerImpl) Error(args ...interface{}) {
	l.mu.RLock()
	l.entry.Error(args...)
	l.mu.RUnlock()
}

func (l *loggerImpl) Errorf(format string, args ...interface{}) {
	l.mu.RLock()
	l.entry.Errorf(format, args...)
	l.mu.RUnlock()
}

func (l *loggerImpl) Fatal(args ...interface{}) {
	l.mu.RLock()
	l.entry.Fatal(args...)
	l.mu.RUnlock()
}

func (l *loggerImpl) Fatalf(format string, args ...interface{}) {
	l.mu.RLock()
	l.entry.Fatalf(format, args...)
	l.mu.RUnlock()
}

func (l *loggerImpl) Panic(args ...interface{}) {
	l.mu.RLock()
	l.entry.Panic(args...)
	l.mu.RUnlock()
}

func (l *loggerImpl) Panicf(format string, args ...interface{}) {
	l.mu.RLock()
	l.entry.Panicf(format, args...)
	l.mu.RUnlock()
}

// 实现Logger接口的字段方法
func (l *loggerImpl) WithField(key string, value interface{}) Logger {
	l.mu.RLock()
	newEntry := l.entry.WithField(key, value)
	l.mu.RUnlock()
	return &loggerImpl{entry: newEntry, file: l.file}
}

func (l *loggerImpl) WithFields(fields map[string]interface{}) Logger {
	l.mu.RLock()
	newEntry := l.entry.WithFields(fields)
	l.mu.RUnlock()
	return &loggerImpl{entry: newEntry, file: l.file}
}

func (l *loggerImpl) WithError(err error) Logger {
	l.mu.RLock()
	newEntry := l.entry.WithError(err)
	l.mu.RUnlock()
	return &loggerImpl{entry: newEntry, file: l.file}
}

// 实现Logger接口的配置方法
func (l *loggerImpl) SetLevel(level logrus.Level) {
	l.mu.Lock()
	l.entry.Logger.SetLevel(level)
	l.mu.Unlock()
}

func (l *loggerImpl) GetLevel() logrus.Level {
	l.mu.RLock()
	level := l.entry.Logger.GetLevel()
	l.mu.RUnlock()
	return level
}

func (l *loggerImpl) SetOutput(output io.Writer) {
	l.mu.Lock()
	l.entry.Logger.SetOutput(output)
	l.mu.Unlock()
}

func (l *loggerImpl) GetOutput() io.Writer {
	l.mu.RLock()
	output := l.entry.Logger.Out
	l.mu.RUnlock()
	return output
}

func (l *loggerImpl) SetFormatter(formatter logrus.Formatter) {
	l.mu.Lock()
	l.entry.Logger.SetFormatter(formatter)
	l.mu.Unlock()
}

// 实现Logger接口的扩展方法
func (l *loggerImpl) WithComponent(component string) Logger {
	return l.WithField("component", component)
}

func (l *loggerImpl) WithRequestID(requestID string) Logger {
	return l.WithField("request_id", requestID)
}

// 实现Logger接口的资源管理方法
func (l *loggerImpl) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.file != nil {
		return l.file.Close()
	}
	return nil
}
