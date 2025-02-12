#!/bin/bash
NLIP_DIR=$(pwd)
NLIP_TEST_DIR=$NLIP_DIR/test
NLIP_FRONTEND_DIR=$NLIP_DIR/src/frontend
NLIP_BACKEND_DIR=$NLIP_DIR/src/backend
NLIP_STATIC_DIR=$NLIP_BACKEND_DIR/static

# 清理并创建目录
if [ -d $NLIP_TEST_DIR ]; then
    echo "Cleaning $NLIP_TEST_DIR"
    if [ -f $NLIP_TEST_DIR/nlip ]; then
        echo "Removing $NLIP_TEST_DIR/nlip"
        rm -rf $NLIP_TEST_DIR/nlip
    fi
    if [ -d $NLIP_TEST_DIR/data ]; then
        echo "Removing $NLIP_TEST_DIR/data"
        rm -rf $NLIP_TEST_DIR/data
    fi
    if [ -d $NLIP_TEST_DIR/uploads ]; then
        echo "Removing $NLIP_TEST_DIR/uploads"
        rm -rf $NLIP_TEST_DIR/uploads
    fi
    if [ -d $NLIP_TEST_DIR/logs ]; then
        echo "Removing $NLIP_TEST_DIR/logs"
        rm -rf $NLIP_TEST_DIR/logs
    fi
fi
if [ -d $NLIP_STATIC_DIR ]; then
    echo "Cleaning $NLIP_STATIC_DIR"
    rm -rf $NLIP_STATIC_DIR
fi

# 复制前端文件
mkdir -p $NLIP_BACKEND_DIR/static/dist
echo "Copying frontend files to backend"
cp -r $NLIP_FRONTEND_DIR/dist $NLIP_BACKEND_DIR/static

# 编译后端
echo "Building backend"
cd $NLIP_BACKEND_DIR

生成swagger文档
echo "Generating API documentation"
swag init
if [ $? -ne 0 ]; then
    echo "Failed to generate API documentation"
    exit 1
fi

CGO_ENABLED=0 
go build -o $NLIP_TEST_DIR/nlip ./main.go

echo "Build completed successfully!"
