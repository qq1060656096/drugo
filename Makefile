.PHONY: help test testa cover build install clean

# CLI 编译配置
CLI_NAME := drugo
CLI_DIR := cmd/drugo
BUILD_DIR := bin

help: ## 显示帮助信息
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-15s\033[0m %s\n", $$1, $$2}'

test: ## 运行单元测试
	go test -count=1 -v ./...

testa: ## 运行单元测试并检测竞态条件
	go test -count=1 -v -race -bench=. ./...

cover: ## 检查覆盖率
	go test -cover ./...

build: ## 编译 drugo CLI 工具
	@mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(CLI_NAME) ./$(CLI_DIR)
	@echo "Built $(BUILD_DIR)/$(CLI_NAME)"

install: ## 安装 drugo CLI 到 GOPATH/bin
	go install ./$(CLI_DIR)
	@echo "Installed $(CLI_NAME) to $$(go env GOPATH)/bin"

clean: ## 清理编译产物
	@rm -rf $(BUILD_DIR)
	@echo "Cleaned $(BUILD_DIR)"

