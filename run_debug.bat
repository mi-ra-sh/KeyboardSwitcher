@echo off
echo Building debug version...
go build -o KeyboardSwitcher_debug.exe .
if %errorlevel% equ 0 (
    echo Starting KeyboardSwitcher (debug)...
    KeyboardSwitcher_debug.exe
) else (
    echo BUILD FAILED
    pause
)
