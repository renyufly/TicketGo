param(
    [string[]]$Strategies = @('naive', 'pessimistic', 'atomic', 'optimistic'),
    [int]$Runs = 3,
    [int]$Users = 1000,
    [int]$Stock = 100,
    [int]$VUs = 200,
    [string]$LabelSuffix = ''
)

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
if (-not $env:DATABASE_URL -or -not $env:JWT_SECRET) {
    throw 'DATABASE_URL and JWT_SECRET must be available through .env or the process environment.'
}

& make build
if ($LASTEXITCODE -ne 0) { throw 'Building TicketGo failed.' }
& make phase2-tool
if ($LASTEXITCODE -ne 0) { throw 'Building phase2lab failed.' }
if (-not (Test-Path -LiteralPath '.tools\k6\k6.exe')) {
    & powershell -NoProfile -ExecutionPolicy Bypass -File scripts/bootstrap-k6.ps1
}

$rawDirectory = Join-Path $projectRoot 'docs\results\phase2\raw'
$generatedDirectory = Join-Path $projectRoot 'tests\load\.generated'
New-Item -ItemType Directory -Force -Path $rawDirectory, $generatedDirectory | Out-Null

foreach ($strategy in $Strategies) {
    if ($strategy -notin @('naive', 'pessimistic', 'atomic', 'optimistic')) {
        throw "Unsupported strategy: $strategy"
    }
    for ($run = 1; $run -le $Runs; $run++) {
        $label = "$strategy$LabelSuffix-run$run"
        $dataset = Join-Path $generatedDirectory "$label-users.json"
        & .\bin\phase2lab.exe prepare --users $Users --stock $Stock --output $dataset
        if ($LASTEXITCODE -ne 0) { throw "Preparing $label failed." }

        $env:SECKILL_INVENTORY_STRATEGY = $strategy
        $env:SECKILL_NAIVE_DELAY = if ($strategy -eq 'naive') { '75ms' } else { '0s' }
        $env:STRATEGY = $strategy
        $env:PHASE2_DATASET = $dataset
        $env:VUS = $VUs.ToString()
        $env:ITERATIONS = $Users.ToString()
        $env:K6_SUMMARY_PATH = Join-Path $rawDirectory "$label-k6.json"

        $stdout = Join-Path $rawDirectory "$label-server.log"
        $stderr = Join-Path $rawDirectory "$label-server-error.log"
        $server = Start-Process -FilePath '.\bin\ticketgo.exe' -PassThru -WindowStyle Hidden -RedirectStandardOutput $stdout -RedirectStandardError $stderr
        $stopFile = Join-Path $generatedDirectory "$label-monitor.stop"
        $monitorOutput = Join-Path $rawDirectory "$label-resource.csv"
        if (Test-Path -LiteralPath $stopFile) { Remove-Item -LiteralPath $stopFile }
        $monitor = $null
        $dockerStats = $null
        try {
            $ready = $false
            for ($attempt = 0; $attempt -lt 50; $attempt++) {
                try {
                    $response = Invoke-WebRequest -UseBasicParsing -Uri 'http://127.0.0.1:8080/health/ready' -TimeoutSec 1
                    if ($response.StatusCode -eq 200) { $ready = $true; break }
                } catch {}
                Start-Sleep -Milliseconds 200
            }
            if (-not $ready) { throw "TicketGo did not become ready for $label." }
            $monitor = Start-Process -FilePath '.\bin\phase2lab.exe' -ArgumentList @('monitor', '--output', $monitorOutput, '--stop-file', $stopFile, '--interval', '50ms') -PassThru -WindowStyle Hidden
            $dockerStatsOutput = Join-Path $rawDirectory "$label-docker-stats.log"
            $dockerStatsError = Join-Path $rawDirectory "$label-docker-stats-error.log"
            $dockerStats = Start-Process -FilePath 'docker.exe' -ArgumentList @('stats', 'ticketgo-postgres-1', '--format', '{{.CPUPerc}}') -PassThru -WindowStyle Hidden -RedirectStandardOutput $dockerStatsOutput -RedirectStandardError $dockerStatsError
            Start-Sleep -Seconds 1
            & .\.tools\k6\k6.exe run tests/load/phase2-seckill.js
            if ($LASTEXITCODE -ne 0) { throw "k6 failed for $label." }
            Start-Sleep -Seconds 1
        } finally {
            New-Item -ItemType File -Force -Path $stopFile | Out-Null
            if ($monitor -and -not $monitor.HasExited) { $monitor.WaitForExit(5000) | Out-Null }
            if ($dockerStats -and -not $dockerStats.HasExited) { Stop-Process -Id $dockerStats.Id }
            if (-not $server.HasExited) { Stop-Process -Id $server.Id }
            $server.WaitForExit(5000) | Out-Null
        }

        $expectation = if ($strategy -eq 'naive') { 'oversold' } else { 'safe' }
        $verificationOutput = Join-Path $rawDirectory "$label-verification.json"
        & .\bin\phase2lab.exe verify --dataset $dataset --expect $expectation --output $verificationOutput
        if ($LASTEXITCODE -ne 0) { throw "Correctness verification failed for $label." }
    }
}

Write-Output "Phase 2 load runs completed. Raw results: $rawDirectory"
