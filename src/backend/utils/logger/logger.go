package logger

import (
    "fmt"
    "log"
    "os"
    "path/filepath"
    "runtime"
    "strings"
    "time"
)

var (
    debugLogger   *log.Logger
    infoLogger    *log.Logger
    warningLogger *log.Logger
    errorLogger   *log.Logger
    logFile       *os.File
)

const (
    logDir     = "logs"
    timeFormat = "2006-01-02"
    logTimeFormat = "2006-01-02 15:04:05"
)

var appEnv string = "development"

func init() {
    // 创建日志目录
    if err := os.MkdirAll(logDir, 0755); err != nil {
        panic(fmt.Sprintf("创建日志目录失败: %v", err))
    }

    // 打开或创建日志文件
    logPath := filepath.Join(logDir, fmt.Sprintf("nlip_%s.log", time.Now().Format(timeFormat)))
    file, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
    if err != nil {
        panic(fmt.Sprintf("打开日志文件失败: %v", err))
    }
    logFile = file

    // 初始化日志记录器
    flags := 0
    debugLogger = log.New(file, "[DEBUG] ", flags)
    infoLogger = log.New(file, "[INFO] ", flags)
    warningLogger = log.New(file, "[WARN] ", flags)
    errorLogger = log.New(file, "[ERROR] ", flags)

    // 启动日志轮转
    go rotateLogDaily()
}

func SetAppEnv(env string) {
    appEnv = env
    if appEnv == "development" {
        Info("appEnv: %s", appEnv)
    } else {
        Info("appEnv: %s, 禁用日志DEBUG输出", appEnv)
    }
}

// Close 关闭日志文件
func Close() {
    if logFile != nil {
        logFile.Close()
    }
}

// rotateLogDaily 每天轮转日志文件
func rotateLogDaily() {
    for {
        now := time.Now()
        next := now.Add(24 * time.Hour)
        next = time.Date(next.Year(), next.Month(), next.Day(), 0, 0, 0, 0, next.Location())
        duration := next.Sub(now)

        timer := time.NewTimer(duration)
        <-timer.C

        // 关闭当前日志文件
        logFile.Close()

        // 创建新的日志文件
        logPath := filepath.Join(logDir, fmt.Sprintf("nlip_%s.log", time.Now().Format(timeFormat)))
        file, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
        if err != nil {
            fmt.Printf("创建新日志文件失败: %v\n", err)
            continue
        }

        // 更新日志记录器
        logFile = file
        flags := 0
        debugLogger = log.New(file, "[DEBUG] ", flags)
        infoLogger = log.New(file, "[INFO] ", flags)
        warningLogger = log.New(file, "[WARN] ", flags)
        errorLogger = log.New(file, "[ERROR] ", flags)

        // 清理旧日志文件（保留30天）
        cleanOldLogs(30)
    }
}

// cleanOldLogs 清理指定天数之前的日志文件
func cleanOldLogs(days int) {
    cutoff := time.Now().AddDate(0, 0, -days)
    files, err := os.ReadDir(logDir)
    if err != nil {
        fmt.Printf("读取日志目录失败: %v\n", err)
        return
    }

    for _, file := range files {
        if !file.IsDir() && strings.HasPrefix(file.Name(), "nlip_") {
            filePath := filepath.Join(logDir, file.Name())
            info, err := file.Info()
            if err != nil {
                continue
            }

            if info.ModTime().Before(cutoff) {
                if err := os.Remove(filePath); err != nil {
                    fmt.Printf("删除旧日志文件失败 %s: %v\n", filePath, err)
                }
            }
        }
    }
}

// getCallerInfo 获取调用者信息
func getCallerInfo() string {
    _, file, line, ok := runtime.Caller(2)
    if !ok {
        return ""
    }
    
    // 查找并截取 src/backend 之后的路径
    if idx := strings.Index(file, "src/backend"); idx != -1 {
        file = file[idx:]
    }
    
    return fmt.Sprintf("%s:%d", file, line)
}

// Debug 记录调试日志
func Debug(format string, v ...interface{}) {
    if appEnv == "development" {
        msg := fmt.Sprintf(format, v...)
        timeStr := time.Now().Format(logTimeFormat)
        debugLogger.Printf("%s %s %s", timeStr, getCallerInfo(), msg)
    }
}

// Info 记录信息日志
func Info(format string, v ...interface{}) {
    msg := fmt.Sprintf(format, v...)
    timeStr := time.Now().Format(logTimeFormat)
    infoLogger.Printf("%s %s %s", timeStr, getCallerInfo(), msg)
}

// Warning 记录警告日志
func Warning(format string, v ...interface{}) {
    msg := fmt.Sprintf(format, v...)
    timeStr := time.Now().Format(logTimeFormat)
    warningLogger.Printf("%s %s %s", timeStr, getCallerInfo(), msg)
}

// Error 记录错误日志
func Error(format string, v ...interface{}) {
    msg := fmt.Sprintf(format, v...)
    timeStr := time.Now().Format(logTimeFormat)
    errorLogger.Printf("%s %s %s", timeStr, getCallerInfo(), msg)
}

// Fatal 记录致命错误并退出
func Fatal(format string, v ...interface{}) {
    msg := fmt.Sprintf(format, v...)
    timeStr := time.Now().Format(logTimeFormat)
    errorLogger.Printf("%s %s FATAL: %s", timeStr, getCallerInfo(), msg)
    os.Exit(1)
}