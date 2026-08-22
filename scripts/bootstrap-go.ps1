$ErrorActionPreference = 'Stop'

$goVersion = '1.27.0'
$archiveSha256 = 'f0c0a0d33ba94f4d2c5dbc887334ce678b21813504ddb3aafcb06e60a5a667c4'
$projectRoot = Split-Path -Parent $PSScriptRoot
$toolsDirectory = Join-Path $projectRoot '.tools'
$goExecutable = Join-Path $toolsDirectory 'go\bin\go.exe'
$archive = Join-Path $toolsDirectory "go$goVersion.windows-amd64.zip"

if (Test-Path -LiteralPath $goExecutable) {
    $installedVersion = & $goExecutable version
    if ($installedVersion -match "go$([regex]::Escape($goVersion))") {
        Write-Output "Project-local $installedVersion is already available."
        exit 0
    }
    throw "A different project-local Go exists at $goExecutable. Remove .tools/go explicitly before changing versions."
}

New-Item -ItemType Directory -Force -Path $toolsDirectory | Out-Null
try {
    & curl.exe --fail --location --output $archive "https://go.dev/dl/go$goVersion.windows-amd64.zip"
    if ($LASTEXITCODE -ne 0) {
        throw "Downloading Go $goVersion failed with exit code $LASTEXITCODE."
    }
    $actualSha256 = (Get-FileHash -Algorithm SHA256 -LiteralPath $archive).Hash.ToLowerInvariant()
    if ($actualSha256 -ne $archiveSha256) {
        throw "Go archive checksum mismatch. Expected $archiveSha256, got $actualSha256."
    }
    Expand-Archive -LiteralPath $archive -DestinationPath $toolsDirectory
} finally {
    if (Test-Path -LiteralPath $archive) {
        Remove-Item -LiteralPath $archive
    }
}

& $goExecutable version

