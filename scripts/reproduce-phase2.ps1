$ErrorActionPreference = 'Stop'
$projectRoot = Split-Path -Parent $PSScriptRoot
Set-Location $projectRoot

if (-not (Test-Path -LiteralPath '.env')) {
    throw 'Missing .env. Copy .env.example to .env and replace JWT_SECRET before reproducing Phase 2.'
}

$listener = Get-NetTCPConnection -LocalPort 8080 -State Listen -ErrorAction SilentlyContinue
if ($listener) {
    throw 'Port 8080 is already in use. Stop the manually running TicketGo service before Phase 2 reproduction; the load runner starts an isolated service for each strategy.'
}

& docker compose up -d
if ($LASTEXITCODE -ne 0) { throw 'Starting PostgreSQL failed.' }
& make migrate-up
if ($LASTEXITCODE -ne 0) { throw 'Applying migrations failed.' }

# Formal same-environment comparison: four strategies, three runs each.
& powershell -NoProfile -ExecutionPolicy Bypass -File scripts/run-phase2-load.ps1 -Runs 3 -Users 1000 -Stock 100 -VUs 200
if ($LASTEXITCODE -ne 0) { throw 'Formal 1000 buyers / 100 tickets comparison failed.' }

# Required optimistic-lock low-contention control. Hotspot data comes from the
# formal comparison above; one VU provides a true zero-conflict baseline.
& powershell -NoProfile -ExecutionPolicy Bypass -File scripts/run-phase2-load.ps1 -Strategies optimistic -Runs 3 -Users 100 -Stock 100 -VUs 1 -LabelSuffix low
if ($LASTEXITCODE -ne 0) { throw 'Optimistic low-contention comparison failed.' }

& powershell -NoProfile -ExecutionPolicy Bypass -File scripts/summarize-phase2-results.ps1
if ($LASTEXITCODE -ne 0) { throw 'Summarizing Phase 2 load results failed.' }

& powershell -NoProfile -ExecutionPolicy Bypass -File scripts/run-phase2-postgresql-labs.ps1
if ($LASTEXITCODE -ne 0) { throw 'PostgreSQL internals and index experiments failed.' }

Write-Output 'Phase 2 reproduction completed: formal load comparison, optimistic low-contention control, summary, and PostgreSQL labs.'
