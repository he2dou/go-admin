.PHONY: all build run tool swagger wire clean docker-build docker-aliyun-login docker-aliyun-tag docker-aliyun-push docker_aliyun_build help

BINARY="draining"

VERSION = $(shell git describe --tags)

all: tool build

line: tool swagger wire docker-build docker-aliyun-tag docker-aliyun-push

install:
	go install github.com/google/wire/cmd/wire@v0.5.0
	go install github.com/swaggo/swag/cmd/swag@v1.8.2

# CGO_ENABLE=0 GOOS=linux GOARCH=amd64 go build -o ${BINARY}
# 可以在编译的时候加上-ldflags "-s -w"参数去掉符号表和调试信息，一般能减小20%的大小

build:
	go build -ldflags "-X main.Version=${VERSION}" -o ${BINARY}

linux:
	SET CGO_ENABLE=0
	SET GOOS=linux
	SET GOARCH=amd64
	go build -ldflags "-X main.Version=${VERSION}" -o ${BINARY}

run:
	@go run .

tool:
	@go fmt ./...
	@go vet ./...

swagger:
	swag init --parseDependency --generalInfo ./docs/swagger/swagger.go --output ./docs/swagger

wire:
	wire gen ./internal/injector

clean:
	@if [ -f ${BINARY} ] ; then rm ${BINARY} ; fi

docker-build:
	docker build -f ./build/Dockerfile -t push --build-arg version=${VERSION} .

docker-aliyun-login:
	docker login --username=mapbridge@1863670751875134 registry.cn-hangzhou.aliyuncs.com

docker-aliyun-base:
	docker tag authbase registry.cn-hangzhou.aliyuncs.com/lomark/push

docker-aliyun-tag:
	docker tag push registry.cn-hangzhou.aliyuncs.com/lomark/push:${VERSION}
	docker tag push registry.cn-hangzhou.aliyuncs.com/lomark/push

docker-aliyun-push:
	docker push registry.cn-hangzhou.aliyuncs.com/lomark/push:${VERSION}
	docker push registry.cn-hangzhou.aliyuncs.com/lomark/push

docker-aliyun-pull:
	docker pull registry.cn-hangzhou.aliyuncs.com/lomark/push

docker_aliyun_build:
	@echo "go mod tidy"
	chmod 755 swag && ./swag init --parseDependency --generalInfo ./docs/swagger/swagger.go --output ./docs/swagger
	chmod 755 wire && ./wire gen ./internal/injector
	go build -ldflags "-X main.Version=${VERSION}" -o ${BINARY}

help:
	@echo "make - 格式化 Go 代码, 并编译生成二进制文件"
	@echo "make build - 编译 Go 代码, 生成二进制文件"
	@echo "make run - 直接运行 Go 代码"
	@echo "make clean - 移除二进制文件和 vim swap files"
	@echo "make tool - 运行 Go 工具 'fmt' and 'vet'"
	@echo "make wire - 生成依赖注入"
	@echo "make swagger - 生成swagger文档"
	@echo "make docker-build - 编译docker镜像"
	@echo "make docker-aliyun-login - 登录阿里云 Docker Registry"
	@echo "make docker-aliyun-tag - 制作阿里云镜像"
	@echo "make docker-aliyun-push - 推送镜像到阿里云仓库"
	@echo "make docker-aliyun-pull -  从阿里云拉取当前的镜像"

