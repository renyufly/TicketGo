$ErrorActionPreference = 'Stop'

$k6Version = 'v2.2.0'
$archiveSha256 = 'ceb2b1e1cf9dbe1303c6c33ec83ffda86dda5c610b4def92064d3c7ebae8d9f4'
$projectRoot = Split-Path -Parent $PSScriptRoot
$toolsDirectory = Join-Path $projectRoot '.tools'
$k6Directory = Join-Path $toolsDirectory 'k6'
$k6Executable = Join-Path $k6Directory 'k6.exe'
$archive = Join-Path $toolsDirectory "k6-$k6Version-windows-amd64.zip"
$extractDirectory = Join-Path $toolsDirectory 'k6-extract'

if (Test-Path -LiteralPath $k6Executable) {
    $installedVersion = & $k6Executable version
    if ($installedVersion -match "k6 $([regex]::Escape($k6Version))") {
        Write-Output "Project-local $installedVersion is already available."
        exit 0
    }
    throw "A different project-local k6 exists at $k6Executable. Remove .tools/k6 explicitly before changing versions."
}

New-Item -ItemType Directory -Force -Path $toolsDirectory | Out-Null
try {
    & curl.exe --fail --location --output $archive "https://github.com/grafana/k6/releases/download/$k6Version/k6-$k6Version-windows-amd64.zip"
    if ($LASTEXITCODE -ne 0) {
        throw "Downloading k6 $k6Version failed with exit code $LASTEXITCODE."
    }
    $actualSha256 = (Get-FileHash -Algorithm SHA256 -LiteralPath $archive).Hash.ToLowerInvariant()
    if ($actualSha256 -ne $archiveSha256) {
        throw "k6 archive checksum mismatch. Expected $archiveSha256, got $actualSha256."
    }
    New-Item -ItemType Directory -Force -Path $extractDirectory | Out-Null
    Expand-Archive -LiteralPath $archive -DestinationPath $extractDirectory
    New-Item -ItemType Directory -Force -Path $k6Directory | Out-Null
    $extractedExecutable = Get-ChildItem -LiteralPath $extractDirectory -Filter k6.exe -Recurse | Select-Object -First 1
    if ($null -eq $extractedExecutable) {
        throw 'The downloaded k6 archive did not contain k6.exe.'
    }
    Copy-Item -LiteralPath $extractedExecutable.FullName -Destination $k6Executable
} finally {
    if (Test-Path -LiteralPath $archive) {
        Remove-Item -LiteralPath $archive
    }
    if (Test-Path -LiteralPath $extractDirectory) {
        Remove-Item -LiteralPath $extractDirectory -Recurse
    }
}

& $k6Executable version

