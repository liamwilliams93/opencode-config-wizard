@echo off
echo Building OpenCode Config Wizard for Windows...
go build -o opencode-config-wizard.exe .
if %errorlevel% neq 0 (
    echo Build failed!
    exit /b 1
)
echo Build successful: opencode-config-wizard.exe
