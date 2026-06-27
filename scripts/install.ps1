# SPDX-License-Identifier: AGPL-3.0-or-later
# Copyright (C) 2026 Kinn Coelho Juliao <kinncj@gmail.com>
#
# Install Heimdall binaries from GitHub Releases on Windows — install only what
# each machine needs.
#
#   irm https://raw.githubusercontent.com/kinncj/Heimdall/main/scripts/install.ps1 | iex
#   # or, with arguments:
#   .\install.ps1 -Components daemon,helper
#   .\install.ps1 -InstallLocation 'C:\tools\heimdall' -Components dashboard
#
# Binaries install to %LOCALAPPDATA%\Heimdall\bin by default; override with
# -InstallLocation.

param(
    [string[]]$Components = @('dashboard'),
    [string]$InstallLocation = "$env:LOCALAPPDATA\Heimdall\bin",
    [string]$Version = 'latest',
    [string]$Repo = 'kinncj/Heimdall'
)

$ErrorActionPreference = 'Stop'
$headers = @{ 'User-Agent' = 'heimdall-install' }
$valid = @('hub', 'dashboard', 'daemon', 'helper')

$arch = if ($env:PROCESSOR_ARCHITECTURE -eq 'ARM64') { 'arm64' } else { 'amd64' }

if ($Version -eq 'latest') {
    $rel = Invoke-RestMethod "https://api.github.com/repos/$Repo/releases/latest" -Headers $headers
    $Version = $rel.tag_name
}
if (-not $Version) { throw "could not resolve the latest release for $Repo" }
Write-Host "install: $Repo $Version (windows/$arch)"

$base = "https://github.com/$Repo/releases/download/$Version"
New-Item -ItemType Directory -Force -Path $InstallLocation | Out-Null

# --- checksums (best effort) ----------------------------------------------
$sums = @{}
try {
    $raw = Invoke-WebRequest "$base/SHA256SUMS" -UseBasicParsing -Headers $headers
    foreach ($line in ($raw.Content -split "`n")) {
        $f = ($line.Trim() -split '\s+')
        if ($f.Count -eq 2) { $sums[$f[1]] = $f[0].ToLower() }
    }
} catch {
    Write-Host "  (no SHA256SUMS published; skipping verification)"
}

foreach ($c in $Components) {
    if ($valid -notcontains $c) { throw "unknown component: $c (want $($valid -join '|'))" }
    $asset = "heimdall-${c}_windows_${arch}.exe"
    $tmp = Join-Path $env:TEMP $asset
    $dest = Join-Path $InstallLocation "heimdall-$c.exe"

    Write-Host "downloading $asset"
    Invoke-WebRequest "$base/$asset" -OutFile $tmp -UseBasicParsing -Headers $headers

    if ($sums.ContainsKey($asset)) {
        $got = (Get-FileHash $tmp -Algorithm SHA256).Hash.ToLower()
        if ($got -ne $sums[$asset]) { throw "checksum mismatch for $asset" }
        Write-Host "  verified $asset"
    }

    Move-Item -Force $tmp $dest
    Write-Host "  installed $dest"
}

if ($env:Path -notlike "*$InstallLocation*") {
    Write-Host ""
    Write-Host "Add $InstallLocation to your PATH:"
    Write-Host "  setx PATH `"$InstallLocation;`$env:PATH`""
}
Write-Host ""
Write-Host "Done. Try:  heimdall-$($Components[0]) --help"
