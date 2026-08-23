$ErrorActionPreference = 'Stop'
$projectRoot = Split-Path -Parent $PSScriptRoot
Set-Location $projectRoot
if (Test-Path -LiteralPath '.env') {
    foreach ($line in Get-Content -LiteralPath '.env') {
        if ($line -match '^\s*([^#][^=]+)=(.*)$') {
            $name = $matches[1].Trim()
            if (-not [Environment]::GetEnvironmentVariable($name, 'Process')) {
                [Environment]::SetEnvironmentVariable($name, $matches[2].Trim(), 'Process')
            }
        }
    }
}
$resultDirectory = Join-Path $projectRoot 'docs\results\phase2'
New-Item -ItemType Directory -Force -Path $resultDirectory | Out-Null

& make phase2-tool
if ($LASTEXITCODE -ne 0) { throw 'Building phase2lab failed.' }
& .\bin\phase2lab.exe internals --output (Join-Path $resultDirectory 'postgresql-internals.json')
if ($LASTEXITCODE -ne 0) { throw 'PostgreSQL internals experiment failed.' }

$indexSQL = Get-Content -Raw -LiteralPath 'scripts\sql\phase2-order-index.sql'
$previousPreference = $ErrorActionPreference
$ErrorActionPreference = 'Continue'
$indexOutput = $indexSQL | & docker compose exec -T postgres psql -U ticketgo -d ticketgo -X -f - 2>&1
$indexExitCode = $LASTEXITCODE
$ErrorActionPreference = $previousPreference
$indexOutput | Set-Content -LiteralPath (Join-Path $resultDirectory 'order-index-explain.txt')
if ($indexExitCode -ne 0) { throw 'Order index experiment failed.' }

$typesSQL = Get-Content -Raw -LiteralPath 'scripts\sql\phase2-index-types.sql'
$ErrorActionPreference = 'Continue'
$typesOutput = $typesSQL | & docker compose exec -T postgres psql -U ticketgo -d ticketgo -X -f - 2>&1
$typesExitCode = $LASTEXITCODE
$ErrorActionPreference = $previousPreference
$typesOutput | Set-Content -LiteralPath (Join-Path $resultDirectory 'index-types-explain.txt')
if ($typesExitCode -ne 0) { throw 'PostgreSQL index type experiment failed.' }

Write-Output "PostgreSQL lab results: $resultDirectory"
