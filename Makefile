.PHONY: all build server client client_v2 clean fmt vet test help

# 编译输出目录
OUTPUT_DIR := bin

# 默认目标
all: server client client_v2

# 编译所有（与 all 相同，提供更直观的命令）
build: all

# 编译 server (注入版本信息)
server:
	@mkdir -p $(OUTPUT_DIR)
	@cd server && \
	GIT_COMMIT=`git rev-parse --short HEAD 2>/dev/null || echo "unknown"` && \
	GIT_TIME=`git log -1 --format=%ct 2>/dev/null || echo "0"` && \
	BUILD_TIME=`date -u +%Y-%m-%dT%H:%M:%SZ` && \
	go build -ldflags "-X main.GitCommit=$$GIT_COMMIT -X main.CommitTime=$$GIT_TIME -X main.BuildTime=$$BUILD_TIME" -o ../$(OUTPUT_DIR)/server ./cmd/main.go
	@echo "✅ server 编译完成: $(OUTPUT_DIR)/server"

# 编译 client
client:
	@mkdir -p $(OUTPUT_DIR)
	@cd client && go build -o ../$(OUTPUT_DIR)/client ./client.go
	@echo "✅ client 编译完成: $(OUTPUT_DIR)/client"

# 编译 client_v2
client_v2:
	@mkdir -p $(OUTPUT_DIR)
	@cd client_v2 && go build -o ../$(OUTPUT_DIR)/client_v2 ./client.go
	@echo "✅ client_v2 编译完成: $(OUTPUT_DIR)/client_v2"

# 清理构建产物
clean:
	@rm -rf $(OUTPUT_DIR)
	@echo "🗑️ 清理完成"

# 格式化代码
fmt:
	@echo "📝 格式化 server 代码..."
	@cd server && go fmt ./...
	@echo "📝 格式化 client 代码..."
	@cd client && go fmt ./...
	@echo "📝 格式化 client_v2 代码..."
	@cd client_v2 && go fmt ./...
	@echo "✅ 格式化完成"

# 代码静态检查
vet:
	@echo "🔍 检查 server 代码..."
	@cd server && go vet ./...
	@echo "🔍 检查 client 代码..."
	@cd client && go vet ./...
	@echo "🔍 检查 client_v2 代码..."
	@cd client_v2 && go vet ./...
	@echo "✅ 检查完成"

# 运行测试
test:
	@echo "🧪 运行 server 测试..."
	@cd server && go test ./...
	@echo "✅ 测试完成"

# 显示帮助
help:
	@echo "📋 可用命令:"
	@echo "  make          - 编译所有模块 (server, client, client_v2)"
	@echo "  make server   - 编译 server"
	@echo "  make client   - 编译 client"
	@echo "  make client_v2 - 编译 client_v2"
	@echo "  make clean    - 清理构建产物"
	@echo "  make fmt      - 格式化所有代码"
	@echo "  make vet      - 代码静态检查"
	@echo "  make test     - 运行测试"
	@echo "  make help     - 显示此帮助信息"
