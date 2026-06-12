param(
    [string]$EnvFile = "deploy/compose/core/.env",
    [string]$Container = "postgres-core"
)

$ErrorActionPreference = "Stop"

if (-not (Test-Path $EnvFile)) {
    throw "Env file not found: $EnvFile"
}

$envs = @{}
Get-Content $EnvFile | ForEach-Object {
    $line = $_.Trim()
    if ($line -eq "" -or $line.StartsWith("#")) {
        return
    }
    $parts = $line.Split("=", 2)
    if ($parts.Count -eq 2) {
        $envs[$parts[0]] = $parts[1]
    }
}

$required = @(
    "POSTGRES_USER",
    "POSTGRES_PASSWORD",
    "POSTGRES_DB",
    "AUDIT_POSTGRES_USER",
    "AUDIT_POSTGRES_PASSWORD",
    "AUDIT_POSTGRES_DB"
)

foreach ($key in $required) {
    if (-not $envs.ContainsKey($key) -or $envs[$key] -eq "") {
        throw "Missing $key in $EnvFile"
    }
}

$sql = @'
SELECT format('CREATE ROLE %I LOGIN PASSWORD %L', :'audit_user', :'audit_password')
WHERE NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = :'audit_user') \gexec

SELECT format('ALTER ROLE %I WITH LOGIN PASSWORD %L', :'audit_user', :'audit_password')
WHERE EXISTS (SELECT 1 FROM pg_roles WHERE rolname = :'audit_user') \gexec

SELECT format('CREATE DATABASE %I OWNER %I', :'audit_db', :'audit_user')
WHERE NOT EXISTS (SELECT 1 FROM pg_database WHERE datname = :'audit_db') \gexec

SELECT format('ALTER DATABASE %I OWNER TO %I', :'audit_db', :'audit_user')
WHERE EXISTS (SELECT 1 FROM pg_database WHERE datname = :'audit_db') \gexec
'@

$sql | docker exec -i `
    -e "PGPASSWORD=$($envs["POSTGRES_PASSWORD"])" `
    $Container `
    psql `
    --username $envs["POSTGRES_USER"] `
    --dbname $envs["POSTGRES_DB"] `
    -v ON_ERROR_STOP=1 `
    -v "audit_user=$($envs["AUDIT_POSTGRES_USER"])" `
    -v "audit_password=$($envs["AUDIT_POSTGRES_PASSWORD"])" `
    -v "audit_db=$($envs["AUDIT_POSTGRES_DB"])"

if ($LASTEXITCODE -ne 0) {
    throw "Failed to ensure audit database in $Container"
}
