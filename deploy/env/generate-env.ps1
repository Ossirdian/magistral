$ErrorActionPreference = "Stop"

$root = Resolve-Path "$PSScriptRoot"
$composeRoot = Resolve-Path (Join-Path $PSScriptRoot "..\compose")
$sourceEnv = Join-Path $root ".env.template"
$targetEnv = Join-Path $root ".env"

Copy-Item $sourceEnv $targetEnv -Force

$vars = @{}
foreach ($line in Get-Content $targetEnv) {
    if ([string]::IsNullOrWhiteSpace($line) -or $line.StartsWith("#")) {
        continue
    }

    $pair = $line -split "=", 2
    if ($pair.Length -eq 2) {
        $vars[$pair[0]] = $pair[1]
    }
}

foreach ($service in @("core", "resources", "saas", "iam", "lesson", "student", "payment", "document", "reporting", "audit", "planning", "frontend", "envoy")) {
    $templatePath = Join-Path $root "$service.env.template"
    if (-not (Test-Path $templatePath)) {
        continue
    }

    $content = Get-Content $templatePath -Raw
    foreach ($key in $vars.Keys) {
        $content = $content.Replace('$' + '{' + $key + '}', $vars[$key])
    }

    $outputDir = Join-Path $composeRoot $service
    New-Item -ItemType Directory -Force $outputDir | Out-Null
    Set-Content -Path (Join-Path $outputDir ".env") -Value $content -NoNewline
}
