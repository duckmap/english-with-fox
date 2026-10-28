@echo off
cd /d "%~dp0"
where go >nul 2>&1
if errorlevel 1 (
 echo Go is not installed. Use PLAY_WINDOWS.bat for the prebuilt game.
 pause
 exit /b 1
)
go run .
if errorlevel 1 pause
