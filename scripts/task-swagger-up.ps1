param(
    [Parameter(Mandatory = $true)]
    [string]$RootDir,

    [string]$Port = "8088",

    [string]$Container = "magistral-swagger"
)

$ErrorActionPreference = "Stop"

$apiRoot = Join-Path $RootDir "shared/api"
$urls = @(
    @{ name = "Resources"; url = "/api/resources/v1/resources.openapi.yaml" },
    @{ name = "SaaS Admin"; url = "/api/saas-admin/v1/saas-admin.openapi.yaml" },
    @{ name = "Lessons"; url = "/api/lessons/v1/lessons.openapi.yaml" },
    @{ name = "Students"; url = "/api/students/v1/students.openapi.yaml" },
    @{ name = "Payments"; url = "/api/payments/v1/payments.openapi.yaml" },
    @{ name = "Documents"; url = "/api/documents/v1/documents.openapi.yaml" },
    @{ name = "Reporting"; url = "/api/reporting/v1/reporting.openapi.yaml" }
) | ConvertTo-Json -Compress

$existingContainer = & docker ps -a --filter "name=^/${Container}$" --format "{{.Names}}"
if ($LASTEXITCODE -ne 0) {
    throw "Failed to inspect Docker containers"
}
if ($existingContainer -eq $Container) {
    & docker rm -f $Container *> $null
    if ($LASTEXITCODE -ne 0) {
        throw "Failed to remove existing Swagger UI container"
    }
}

$runArgs = @(
    "run",
    "-d",
    "--name", $Container,
    "-p", "${Port}:8080",
    "-e", "URLS=$urls",
    "-e", "URLS_PRIMARY_NAME=Resources",
    "-v", "${apiRoot}:/usr/share/nginx/html/api:ro",
    "swaggerapi/swagger-ui"
)

& docker @runArgs
if ($LASTEXITCODE -ne 0) {
    throw "Failed to start Swagger UI container"
}

Write-Host "Swagger UI started: http://localhost:$Port"
