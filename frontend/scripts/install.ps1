$ErrorActionPreference = "Stop"
$npmCmd = Join-Path ${env:ProgramFiles} "nodejs\npm.cmd"
if (-not (Test-Path $npmCmd)) {
    Write-Error "npm.cmd not found at $npmCmd."
    exit 1
}
Set-Location (Join-Path $PSScriptRoot "..")
& $npmCmd install @args
