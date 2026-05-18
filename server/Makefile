.PHONY: all build run test clean fmt vet install

# 项目名称
PROJECT_NAME := gateway
# 可执行文件输出路径
BUILD_DIR := bin
# 主入口文件
MAIN_FILE := ./cmd/main.go
# 配置文件路径
CONFIG_PATH := config/config.yaml

# 默认目标
all: build

# 构建项目
build:
	@mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(PROJECT_NAME) $(MAIN_FILE)
	@echo "Build completed: $(BUILD_DIR)/$(PROJECT_NAME)"

# 运行项目
run: build
	./$(BUILD_DIR)/$(PROJECT_NAME) -config $(CONFIG_PATH)

# 运行项目（带配置参数）
run-dev: build
	./$(BUILD_DIR)/$(PROJECT_NAME) -config config/config.yaml

# 运行测试
test:
	go test -v ./...

# 运行测试（带覆盖率）
test-cover:
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# 清理构建产物
clean:
	rm -rf $(BUILD_DIR)
	rm -f coverage.out coverage.html
	@echo "Clean completed"

# 格式化代码
fmt:
	go fmt ./...
	@echo "Format completed"

# 代码检查
vet:
	go vet ./...
	@echo "Vet completed"

# 安装依赖
install:
	go mod download
	go mod tidy
	@echo "Dependencies installed"

# 查看帮助
help:
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@echo "  all       - 构建项目（默认）"
	@echo "  build     - 构建项目"
	@echo "  run       - 运行项目"
	@echo "  run-dev   - 运行项目（开发模式）"
	@echo "  test      - 运行测试"
	@echo "  test-cover - 运行测试并生成覆盖率报告"
	@echo "  clean     - 清理构建产物"
	@echo "  fmt       - 格式化代码"
	@echo "  vet       - 代码检查"
	@echo "  install   - 安装依赖"
	@echo "  help      - 显示帮助信息"