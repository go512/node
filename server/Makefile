#- 优先从 Git Tag 中获取（git describe --tags），适合打正式版本标签的场景
#- 无 Tag 时用 --always 取最新 Commit 哈希
#- --dirty 标记工作区是否有未提交修改
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "unknown")
#格式为 YYYY-MM-DDTHH:MM:SS（ISO 8601 格式），便于跨平台时间解析
BUILD_TIME := $(shell date +%Y-%m-%dT%H:%M:%S)
#- 取当前代码的 Git Commit 短 ID（7 位），无 Git 仓库时默认 unknown
COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")

#通过 -X 选项将编译时变量注入到 Go 程序的指定包中，格式为 -X 包名.变量名=值
LDFLAGS := -X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME) -X main.Commit=$(COMMIT)
# 可选：添加编译优化（移除调试信息，减小二进制体积）
LDFLAGS += -s -w


outdir:
	mkdir -p ./output/

version:
	@echo "Version: $(VERSION)"
	@echo "BuildTime: $(BUILD_TIME)"
	@echo "Commit: $(COMMIT)"

build: outdir
	go build -ldflags "$(LDFLAGS)" -o ./output/ ./cmd
