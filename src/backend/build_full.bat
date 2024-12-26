@echo off
:: Set console to UTF-8
chcp 65001 > nul

:: Save current directory
set "INITIAL_DIR=%CD%"

echo Starting full build process...

:: Switch to frontend directory
echo Switching to frontend directory...
cd ..\frontend

:: Build frontend
echo Building frontend...
call pnpm run build

if errorlevel 1 (
    echo Frontend build failed!
    cd "%INITIAL_DIR%"
    pause
    exit /b 1
)

:: Handle static directory
echo Processing static directory...
if exist "..\backend\static" (
    echo Removing old static directory...
    rd /s /q "..\backend\static"
)
mkdir "..\backend\static"

:: Copy dist folder
echo Copying dist folder to static directory...
xcopy "dist" "..\backend\static\dist\" /E /I /H /Y

:: Switch to backend directory
echo Switching to backend directory...
cd ..\backend

:: Build backend
echo Building backend...
call goreleaser release --snapshot --skip=publish --clean

if errorlevel 1 (
    echo Backend build failed!
    cd "%INITIAL_DIR%"
    pause
    exit /b 1
)

:: Create target directory if not exists
if not exist "D:\Downloads\nlip" mkdir "D:\Downloads\nlip"

:: Copy executable
echo Copying nlip.exe to target directory...
copy /Y "dist\nlip_windows_amd64_v1\nlip.exe" "D:\Downloads\nlip\"

echo Build process completed!

:: Run the program
echo Starting nlip...
cd "D:\Downloads\nlip"
start nlip.exe

:: Return to initial directory
cd "%INITIAL_DIR%"
