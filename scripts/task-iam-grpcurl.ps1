param(
    [Parameter(Mandatory = $true)]
    [string]$Grpcurl,

    [Parameter(Mandatory = $true)]
    [string]$Method,

    [string]$Data = "",

    [string]$Addr = "localhost:50053"
)

$ErrorActionPreference = "Stop"
$Grpcurl = (Resolve-Path $Grpcurl).Path

if ([string]::IsNullOrWhiteSpace($Data)) {
    & $Grpcurl -plaintext $Addr $Method
    exit $LASTEXITCODE
}

& $Grpcurl -plaintext -d $Data $Addr $Method
exit $LASTEXITCODE
