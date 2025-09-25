@echo off
REM Build script for Figaro Go application on Windows
REM This script handles CGO requirements for Windows builds

echo 🏗️ Building Figaro for Windows...

REM Set build variables
if "%VERSION%"=="" set VERSION=dev
for /f "tokens=*" %%i in ('powershell -command "Get-Date -Format 'yyyy-MM-dd HH:mm:ss UTC'"') do set BUILD_TIME=%%i
for /f "tokens=*" %%i in ('git rev-parse --short HEAD 2^>nul') do set GIT_COMMIT=%%i
if "%GIT_COMMIT%"=="" set GIT_COMMIT=unknown

REM Build flags
set LDFLAGS=-s -w -X "main.Version=%VERSION%" -X "main.BuildTime=%BUILD_TIME%" -X "main.GitCommit=%GIT_COMMIT%"

echo 📋 Build Information:
echo    Version: %VERSION%
echo    Platform: windows/amd64
echo    Build Time: %BUILD_TIME%
echo    Git Commit: %GIT_COMMIT%
echo.

REM Clean previous builds
echo 🧹 Cleaning previous builds...
if exist figaro.exe del figaro.exe
if exist figaro del figaro

REM Check for GCC (required for CGO)
gcc --version >nul 2>&1
if %ERRORLEVEL% NEQ 0 (
    echo ❌ Error: GCC not found. CGO requires a C compiler.
    echo.
    echo 💡 Install options:
    echo    - Using Chocolatey: choco install mingw
    echo    - Using winget: winget install mingw-w64
    echo    - Download TDM-GCC: https://jmeubank.github.io/tdm-gcc/
    echo.
    echo 🔄 Alternative: Use CGO-free build with modernc.org/sqlite
    echo    See CGO_FREE_WINDOWS.md for instructions
    exit /b 1
)

REM Download dependencies
echo 📦 Downloading dependencies...
go mod download

REM Run tests
echo 🧪 Running tests...
go test ./...

REM Build with CGO enabled
echo 🔨 Building binary with CGO...
set CGO_ENABLED=1
set CC=gcc
set GOOS=windows
set GOARCH=amd64

go build -ldflags="%LDFLAGS%" -o figaro.exe ./cmd/figaro

if %ERRORLEVEL% NEQ 0 (
    echo ❌ Build failed. Try the CGO-free alternative:
    echo    See CGO_FREE_WINDOWS.md for instructions
    exit /b 1
)

for %%A in (figaro.exe) do set FILE_SIZE=%%~zA

echo ✅ Build completed successfully!
echo    Binary: figaro.exe
echo    Size: %FILE_SIZE% bytes
echo.
echo 🚀 To run the application:
echo    figaro.exe
echo.
echo 🐳 To build Docker image:
echo    docker build -t figaro:%VERSION% .

pause