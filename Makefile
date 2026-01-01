.PHONY: help test cover

help: ## 显示帮助信息
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-15s\033[0m %s\n", $$1, $$2}'

test: ## 运行单元测试
	go test -count=1 -v ./...

testa: ## 运行单元测试并检测竞态条件
	go test -count=1 -v -race bench=. ./...

cover: ## 检查覆盖率
	go test -cover ./...

