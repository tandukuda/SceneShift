@echo off
echo Building SceneShift...
echo.

rsrc -manifest sceneshift.manifest -ico icon.ico -o sceneshift.syso
if %errorlevel% neq 0 (
    echo ❌ Resource compilation failed!
    pause
    exit /b 1
)

go build -o SceneShift.exe
if %errorlevel% neq 0 (
    echo ❌ Build failed!
    pause
    exit /b 1
)

del sceneshift.syso

echo.
echo ✅ Build complete: SceneShift.exe
echo.
pause
