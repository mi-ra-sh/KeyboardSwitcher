@echo off
echo Building KeyboardSwitcher...
go build -ldflags="-H windowsgui -s -w" -o KeyboardSwitcher.exe .
if %errorlevel% equ 0 (
    echo OK: KeyboardSwitcher.exe
) else (
    echo BUILD FAILED
    pause
)
