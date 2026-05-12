param(
    [ValidateSet("up", "down")]
    [string]$Direction = "up",
    [string]$DatabaseUrl = $env:DATABASE_URL
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

if (-not $DatabaseUrl) {
    Write-Error "DATABASE_URL is not set. Pass -DatabaseUrl or set env:DATABASE_URL."
    exit 1
}

$root = Split-Path -Parent $PSScriptRoot
$migrationsDir = Join-Path $root "migrations"

if (-not (Test-Path $migrationsDir)) {
    Write-Error "Migrations directory not found: $migrationsDir"
    exit 1
}

if ($Direction -eq "up") {
    $files = Get-ChildItem -Path $migrationsDir -Filter "*.up.sql" | Sort-Object Name
} else {
    $files = Get-ChildItem -Path $migrationsDir -Filter "*.down.sql" | Sort-Object Name -Descending
}

if (-not $files) {
    Write-Host "No migration files found for direction: $Direction"
    exit 0
}

foreach ($file in $files) {
    Write-Host "Applying $Direction migration: $($file.Name)"
    # Options before --dbname so Windows psql does not treat -v/-f as part of a positional URI.
    & psql -v ON_ERROR_STOP=1 --dbname="$DatabaseUrl" -f "$($file.FullName)"
    if ($LASTEXITCODE -ne 0) {
        Write-Error "Migration failed: $($file.Name)"
        exit $LASTEXITCODE
    }
}

Write-Host "All '$Direction' migrations applied successfully."
