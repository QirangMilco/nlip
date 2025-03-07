@echo off
:: 设置控制台代码页为UTF-8
chcp 65001 > nul

:: 设置目标目录
set "DEST_DIR=..\backend\static"

:: 重定向输出到build.log
(
    echo 开始构建前端项目...
    :: 执行npm构建
    call pnpm run build

    if errorlevel 1 (
        echo 构建失败！
        pause
        exit /b 1
    )

    echo 构建完成，开始移动文件...

    :: 如果目标目录存在，则删除
    if exist "%DEST_DIR%" (
        echo 删除旧的static目录...
        rd /s /q "%DEST_DIR%"
    )

    :: 创建目标目录
    mkdir "%DEST_DIR%"

    :: 复制整个dist文件夹到static目录
    echo 复制dist文件夹到static目录...
    xcopy "dist" "%DEST_DIR%\dist\" /E /I /H /Y

    echo 完成！
) > build.log 2>&1

:: 若未报错，则脚本运行完毕后直接关闭控制台，若报错，则暂停，等待用户输入
if errorlevel 0 (
    exit /b 0
) else (
    :: 显示日志文件内容
    type build.log
    pause
    exit /b 1
)
