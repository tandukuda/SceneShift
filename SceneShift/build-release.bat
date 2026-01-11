@echo off
setlocal enabledelayedexpansion

REM ================================================
REM SceneShift Build Script v2.1
REM ================================================

set VERSION=2.1.0
set /p VERSION="Enter version (default %VERSION%): " || set VERSION=%VERSION%

REM Get current date/time
for /f "tokens=2 delims==" %%a in ('wmic OS Get localdatetime /value') do set "dt=%%a"
set BUILD_DATE=%dt:~0,4%-%dt:~4,2%-%dt:~6,2%
set BUILD_TIME=%dt:~8,2%:%dt:~10,2%:%dt:~12,2%

REM Get git commit (if available)
set GIT_COMMIT=unknown
for /f %%i in ('git rev-parse --short HEAD 2^>nul') do set GIT_COMMIT=%%i

echo.
echo ================================================
echo SceneShift Release Build
echo ================================================
echo Version:    %VERSION%
echo Date:       %BUILD_DATE% %BUILD_TIME%
echo Commit:     %GIT_COMMIT%
echo ================================================
echo.

REM Clean previous builds
if exist sceneshift.syso del sceneshift.syso
if exist SceneShift.exe del SceneShift.exe

echo [1/4] Compiling resources...
rsrc -manifest sceneshift.manifest -ico icon.ico -o sceneshift.syso

if %errorlevel% neq 0 (
    echo.
    echo ❌ ERROR: Resource compilation failed!
    echo.
    echo Make sure rsrc is installed:
    echo    go install github.com/akavel/rsrc@latest
    echo.
    pause
    exit /b 1
)

echo [2/4] Building executable with version info...
go build -ldflags="-s -w -X main.Version=%VERSION% -X main.BuildDate=%BUILD_DATE% -X main.GitCommit=%GIT_COMMIT%" -trimpath -o SceneShift.exe

if %errorlevel% neq 0 (
    echo.
    echo ❌ ERROR: Build failed!
    echo.
    pause
    exit /b 1
)

echo [3/4] Cleaning up temporary files...
del sceneshift.syso

echo [4/4] Creating release package...
if not exist "release" mkdir release
copy /Y SceneShift.exe release\ >nul
if exist README.md copy /Y README.md release\ >nul
if exist LICENSE copy /Y LICENSE release\ >nul

REM Get file size
for %%A in (SceneShift.exe) do set SIZE=%%~zA
set /a SIZE_MB=%SIZE% / 1048576

echo.
echo ================================================
echo ✅ BUILD SUCCESSFUL!
echo ================================================
echo.
echo Output:        release\SceneShift.exe
echo Size:          %SIZE% bytes (~%SIZE_MB% MB)
echo Version:       %VERSION%
echo Build Date:    %BUILD_DATE%
echo Git Commit:    %GIT_COMMIT%
echo.
echo ================================================
echo.
echo Test the build:
echo    cd release
echo    SceneShift.exe --version
echo.
pause
