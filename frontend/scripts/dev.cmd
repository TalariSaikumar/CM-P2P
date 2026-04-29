@echo off
cd /d "%~dp0.."
"%ProgramFiles%\nodejs\node.exe" ".\scripts\run-next.cjs" dev %*
