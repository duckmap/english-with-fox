@echo off
cd /d "%~dp0"
if exist "build\WordGame.exe" (
 start "" "build\WordGame.exe"
) else (
 echo Run: go run .
 pause
)
