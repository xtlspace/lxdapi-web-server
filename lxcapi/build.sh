#!/bin/bash

set -e

echo "开始编译 lxcapi..."

if ! command -v go &> /dev/null; then
    echo "错误: 未找到 Go 编译器"
    exit 1
fi

echo "Go 版本: $(go version)"

echo "清理旧文件..."
rm -f lxcapi-amd64 lxcapi-arm64

echo "下载依赖..."
go mod tidy
go mod download

echo ""
echo "准备嵌入式资源..."
rm -rf cmd/lxcapi/templates cmd/lxcapi/static
cp -r ./lxdweb/templates cmd/lxcapi/
cp -r ./lxdweb/static cmd/lxcapi/
echo "✓ 资源文件已准备"

echo ""
echo "编译 linux/amd64..."
cd cmd/lxcapi
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o "../../lxcapi-amd64" .
cd ../..
echo "✓ lxcapi-amd64 编译完成"

echo ""
echo "编译 linux/arm64..."
cd cmd/lxcapi
GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o "../../lxcapi-arm64" .
cd ../..
echo "✓ lxcapi-arm64 编译完成"

echo ""
echo "编译完成！"
ls -lh lxcapi-* | awk '{print $9 " (" $5 ")"}'

echo ""
echo "清理临时文件..."
rm -rf cmd/lxcapi/templates
echo "✓ 临时文件已清理"

echo ""
echo "开始打包..."

rm -rf release
mkdir -p release

echo "打包 amd64 版本..."
mkdir -p release/lxcapi-amd64
cp lxcapi-amd64 release/lxcapi-amd64/
cp -r configs release/lxcapi-amd64/

# 打包 OpenGFW 插件
mkdir -p release/lxcapi-amd64/plugins/opengfw/bin
mkdir -p release/lxcapi-amd64/plugins/opengfw/data
cp plugins/opengfw/bin/OpenGFW-linux-amd64 release/lxcapi-amd64/plugins/opengfw/bin/
cp plugins/opengfw/data/*.dat release/lxcapi-amd64/plugins/opengfw/data/

# 打包 Nginx 插件
mkdir -p release/lxcapi-amd64/plugins/nginx/conf/sites
mkdir -p release/lxcapi-amd64/plugins/nginx/ssl
cp plugins/nginx/*.tmpl release/lxcapi-amd64/plugins/nginx/ 2>/dev/null || true
cp plugins/nginx/*.md release/lxcapi-amd64/plugins/nginx/ 2>/dev/null || true
cp plugins/nginx/*.sh release/lxcapi-amd64/plugins/nginx/ 2>/dev/null || true
chmod +x release/lxcapi-amd64/plugins/nginx/*.sh 2>/dev/null || true

cd release
tar -czf lxcapi-linux-amd64.tar.gz lxcapi-amd64
rm -rf lxcapi-amd64
cd ..
echo "✓ lxcapi-linux-amd64.tar.gz 打包完成"

echo "打包 arm64 版本..."
mkdir -p release/lxcapi-arm64
cp lxcapi-arm64 release/lxcapi-arm64/
cp -r configs release/lxcapi-arm64/

# 打包 OpenGFW 插件（仅对应架构）
mkdir -p release/lxcapi-arm64/plugins/opengfw/bin
mkdir -p release/lxcapi-arm64/plugins/opengfw/data
cp plugins/opengfw/bin/OpenGFW-linux-arm64 release/lxcapi-arm64/plugins/opengfw/bin/
cp plugins/opengfw/data/*.dat release/lxcapi-arm64/plugins/opengfw/data/

# 打包 Nginx 插件
mkdir -p release/lxcapi-arm64/plugins/nginx/conf/sites
mkdir -p release/lxcapi-arm64/plugins/nginx/ssl
cp plugins/nginx/*.tmpl release/lxcapi-arm64/plugins/nginx/ 2>/dev/null || true
cp plugins/nginx/*.md release/lxcapi-arm64/plugins/nginx/ 2>/dev/null || true
cp plugins/nginx/*.sh release/lxcapi-arm64/plugins/nginx/ 2>/dev/null || true
chmod +x release/lxcapi-arm64/plugins/nginx/*.sh 2>/dev/null || true

cd release
tar -czf lxcapi-linux-arm64.tar.gz lxcapi-arm64
rm -rf lxcapi-arm64
cd ..
echo "✓ lxcapi-linux-arm64.tar.gz 打包完成"

echo ""
echo "所有文件已打包到 release 目录："
ls -lh release/*.tar.gz | awk '{print $9 " (" $5 ")"}'
