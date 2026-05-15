param(
    [Parameter(Mandatory = $true)]
    [string]$Modules,

    [Parameter(Mandatory = $true)]
    [string]$GolangciLint
)

$ErrorActionPreference = "Stop"

$moduleList = $Modules.Split(" ", [System.StringSplitOptions]::RemoveEmptyEntries)
$hasErrors = $false
$configPath = Join-Path $PSScriptRoot '..\\.golangci.yml'

Write-Host "Linting all modules..."

foreach ($module in $moduleList) {
    if (-not (Test-Path $module)) {
        continue
    }

    Write-Host "Linting $module module"

    if (Test-Path $configPath) {
        & $GolangciLint run "$module/..." "--config=$configPath"
    }
    else {
        & $GolangciLint run "$module/..."
    }

    if ($LASTEXITCODE -ne 0) {
        $hasErrors = $true
    }
}

if ($hasErrors) {
    exit 1
}
