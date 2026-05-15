param(
    [Parameter(Mandatory = $true)]
    [string]$Modules,

    [Parameter(Mandatory = $true)]
    [string]$Gofumpt,

    [Parameter(Mandatory = $true)]
    [string]$Gci,

    [Parameter(Mandatory = $true)]
    [string]$ImportPrefix
)

$ErrorActionPreference = "Stop"

$moduleList = $Modules.Split(" ", [System.StringSplitOptions]::RemoveEmptyEntries)

foreach ($module in $moduleList) {
    if (-not (Test-Path $module)) {
        continue
    }

    Write-Host "Formatting $module with gofumpt..."

    $goFiles = Get-ChildItem -Path $module -Recurse -File -Filter *.go |
        Where-Object { $_.FullName -notmatch '[\\/]+mocks[\\/]+' }

    foreach ($file in $goFiles) {
        & $Gofumpt -extra -w $file.FullName
    }

    Write-Host "Sorting imports in $module with gci..."

    foreach ($file in $goFiles) {
        & $Gci write -s standard -s default -s "prefix($ImportPrefix)" $file.FullName
    }
}
