# Runs Next with node.exe directly (avoids next.cmd + cmd.exe PATH issues on Windows / Git Bash).
$ErrorActionPreference = "Stop"
$nodeExe = Join-Path ${env:ProgramFiles} "nodejs\node.exe"
if (-not (Test-Path $nodeExe)) {
    Write-Error "node.exe not found at $nodeExe. Install Node.js or adjust the path in this script."
    exit 1
}
Set-Location (Join-Path $PSScriptRoot "..")
$runner = Join-Path (Get-Location) "scripts\run-next.cjs"
& $nodeExe $runner dev @args
