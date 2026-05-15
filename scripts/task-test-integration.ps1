param(
    [Parameter(Mandatory = $true)]
    [string]$Modules
)

$ErrorActionPreference = "Stop"

foreach ($module in $Modules.Split(" ", [System.StringSplitOptions]::RemoveEmptyEntries)) {
    if (-not (Test-Path $module)) {
        continue
    }

    Push-Location $module
    try {
        go test -tags=integration ./...
    }
    finally {
        Pop-Location
    }
}
