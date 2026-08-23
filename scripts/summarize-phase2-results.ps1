$ErrorActionPreference = 'Stop'
$projectRoot = Split-Path -Parent $PSScriptRoot
Set-Location $projectRoot
$rawDirectory = Join-Path $projectRoot 'docs\results\phase2\raw'
$strategies = @('naive', 'pessimistic', 'atomic', 'optimistic')
$summary = [ordered]@{
    generated_at = (Get-Date).ToUniversalTime().ToString('o')
    scenario = [ordered]@{ users = 1000; stock = 100; requests = 1000; vus = 200; runs = 3 }
    strategies = [ordered]@{}
}

foreach ($strategy in $strategies) {
    $runs = @()
    for ($run = 1; $run -le 3; $run++) {
        $prefix = "$strategy-run$run"
        $k6 = Get-Content -Raw -LiteralPath (Join-Path $rawDirectory "$prefix-k6.json") | ConvertFrom-Json
        $verification = Get-Content -Raw -LiteralPath (Join-Path $rawDirectory "$prefix-verification.json") | ConvertFrom-Json
        $resource = Import-Csv -LiteralPath (Join-Path $rawDirectory "$prefix-resource.csv")
        $runs += [ordered]@{
            run = $run
            qps = [math]::Round($k6.metrics.http_reqs.values.rate, 2)
            p50_ms = [math]::Round($k6.metrics.http_req_duration.values.med, 2)
            p95_ms = [math]::Round($k6.metrics.http_req_duration.values.'p(95)', 2)
            p99_ms = [math]::Round($k6.metrics.http_req_duration.values.'p(99)', 2)
            error_rate = $k6.metrics.http_req_failed.values.rate
            successes = $k6.metrics.seckill_success.values.count
            sold_out = $k6.metrics.seckill_sold_out.values.count
            busy = $k6.metrics.seckill_busy.values.count
            retry_average_all_requests = [math]::Round($k6.metrics.seckill_optimistic_retries.values.avg, 3)
            peak_lock_waiters = [int](($resource | Measure-Object lock_waiters -Maximum).Maximum)
            peak_active_sessions = [int](($resource | Measure-Object active_sessions -Maximum).Maximum)
            peak_oldest_active_transaction_ms = [math]::Round([double](($resource | Measure-Object oldest_active_transaction_ms -Maximum).Maximum), 3)
            available = $verification.available
            sold = $verification.sold
            orders = $verification.orders
            safe = $verification.safe
            oversold = $verification.oversold
        }
    }
    $cpuPath = Join-Path $rawDirectory "$strategy-run1-docker-stats.log"
    if (-not (Test-Path -LiteralPath $cpuPath)) {
        $cpuPath = Join-Path $rawDirectory "$($strategy)cpu-run1-docker-stats.log"
    }
    $cpuRaw = Get-Content -Raw -LiteralPath $cpuPath
    $cpuValues = [regex]::Matches($cpuRaw, '([0-9]+(?:\.[0-9]+)?)%') | ForEach-Object { [double]$_.Groups[1].Value }
    $summary.strategies[$strategy] = [ordered]@{
        mean_qps = [math]::Round(($runs.qps | Measure-Object -Average).Average, 2)
        mean_p50_ms = [math]::Round(($runs.p50_ms | Measure-Object -Average).Average, 2)
        mean_p95_ms = [math]::Round(($runs.p95_ms | Measure-Object -Average).Average, 2)
        mean_p99_ms = [math]::Round(($runs.p99_ms | Measure-Object -Average).Average, 2)
        representative_peak_postgres_cpu_percent = ($cpuValues | Measure-Object -Maximum).Maximum
        mean_peak_lock_waiters = [math]::Round(($runs.peak_lock_waiters | Measure-Object -Average).Average, 2)
        mean_peak_oldest_active_transaction_ms = [math]::Round(($runs.peak_oldest_active_transaction_ms | Measure-Object -Average).Average, 3)
        all_runs_correct = -not ($runs.safe -contains $false)
        all_runs_oversold = -not ($runs.oversold -contains $false)
        runs = $runs
    }
}

$optimisticLow = @()
for ($run = 1; $run -le 3; $run++) {
    $k6 = Get-Content -Raw -LiteralPath (Join-Path $rawDirectory "optimisticlow-run$run-k6.json") | ConvertFrom-Json
    $optimisticLow += [ordered]@{
        run = $run
        qps = [math]::Round($k6.metrics.http_reqs.values.rate, 2)
        p95_ms = [math]::Round($k6.metrics.http_req_duration.values.'p(95)', 2)
        retry_average_all_requests = [math]::Round($k6.metrics.seckill_optimistic_retries.values.avg, 3)
    }
}
$summary.optimistic_low_contention = [ordered]@{
    users = 100
    stock = 100
    vus = 1
    runs = $optimisticLow
}

$output = Join-Path $projectRoot 'docs\results\phase2\concurrency-summary.json'
$summary | ConvertTo-Json -Depth 8 | Set-Content -LiteralPath $output
Write-Output $output
