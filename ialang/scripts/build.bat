@echo off
setlocal EnableExtensions

set "SCRIPT_DIR=%~dp0"
for %%I in ("%SCRIPT_DIR%..") do set "ROOT_DIR=%%~fI"

if "%~1"=="" (
    set "OUTPUT=%ROOT_DIR%\bin\ialang.exe"
) else (
    set "OUTPUT=%~1"
)

if "%~2"=="" (
    set "TARGET=./cmd/ialang"
) else (
    set "TARGET=%~2"
)

for %%I in ("%OUTPUT%") do set "OUT_DIR=%%~dpI"
if not exist "%OUT_DIR%" mkdir "%OUT_DIR%"

pushd "%ROOT_DIR%" >nul
go build -o "%OUTPUT%" "%TARGET%"
if errorlevel 1 (
    echo Build failed.
    popd >nul
    exit /b 1
)
popd >nul

# upx
upx "%OUTPUT%"

echo Build succeeded: %OUTPUT%
exit /b 0
