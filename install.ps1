# ==============================================================================
# Android-MCP-go PowerShell Installer (Windows)
# Repository: https://github.com/tintupratap/Android-MCP-go
# Releases: https://github.com/tintupratap/Android-MCP-go/releases
# Author: Ranapratap (tintupratap@gmail.com)
#
# Usage (PowerShell):
#   irm https://raw.githubusercontent.com/tintupratap/Android-MCP-go/main/install.ps1 | iex
#   $env:VERSION="v0.5.0"; irm https://raw.githubusercontent.com/tintupratap/Android-MCP-go/main/install.ps1 | iex
# ==============================================================================

$ErrorActionPreference = "Stop"

function Write-HostColor($text, $color) {
    Write-Host "[Android-MCP] $text" -ForegroundColor $color
}

Write-HostColor "Starting Android-MCP-go Installation..." "Cyan"

# 1. Detect Architecture
$Arch = $env:PROCESSOR_ARCHITECTURE
switch ($Arch) {
    "AMD64" { $Arch = "amd64" }
    "ARM64" { $Arch = "arm64" }
    Default {
        Write-HostColor "Unsupported architecture: $Arch" "Red"
        exit 1
    }
}

Write-HostColor "Detected Environment: OS=windows, ARCH=$Arch" "Cyan"

# 2. Determine Target Directory
$InstallDir = Join-Path $env:LOCALAPPDATA "Programs\android-mcp"
if (-not (Test-Path $InstallDir)) {
    New-Item -ItemType Directory -Path $InstallDir -Force | Out-Null
}

$BinaryTarget = Join-Path $InstallDir "android-mcp.exe"
$Installed = $false

# 3. Add to User PATH if not present
$UserPath = [Environment]::GetEnvironmentVariable("PATH", "User")
if ($UserPath -notlike "*$InstallDir*") {
    $NewPath = "$UserPath;$InstallDir"
    [Environment]::SetEnvironmentVariable("PATH", $NewPath, "User")
    $env:PATH = "$env:PATH;$InstallDir"
    Write-HostColor "Added $InstallDir to User PATH" "Yellow"
}

# 4. Resolve Release Tag (Supports Pre-releases, Releases, and Git Tags)
if ($env:VERSION) {
    $TargetVer = $env:VERSION
} else {
    Write-HostColor "Querying latest release/pre-release tag from GitHub..." "Cyan"
    try {
        $ReleasesJson = Invoke-RestMethod -Uri "https://api.github.com/repos/tintupratap/Android-MCP-go/releases" -Headers @{"User-Agent"="Android-MCP-Installer"} -ErrorAction Stop
        if ($ReleasesJson -and $ReleasesJson.Count -gt 0) {
            $TargetVer = $ReleasesJson[0].tag_name
        }
    } catch {}

    if (-not $TargetVer) {
        try {
            $TagsJson = Invoke-RestMethod -Uri "https://api.github.com/repos/tintupratap/Android-MCP-go/tags" -Headers @{"User-Agent"="Android-MCP-Installer"} -ErrorAction Stop
            if ($TagsJson -and $TagsJson.Count -gt 0) {
                $TargetVer = $TagsJson[0].name
            }
        } catch {}
    }
}

if ($TargetVer) {
    Write-HostColor "Targeting GitHub Release Tag: $TargetVer" "Cyan"
    $ReleaseUrl = "https://github.com/tintupratap/Android-MCP-go/releases/download/$TargetVer/android-mcp-windows-$Arch.exe"
    $ZipUrl     = "https://github.com/tintupratap/Android-MCP-go/releases/download/$TargetVer/android-mcp-$TargetVer-windows-$Arch.zip"
} else {
    Write-HostColor "Downloading latest prebuilt release from GitHub Releases..." "Cyan"
    $ReleaseUrl = "https://github.com/tintupratap/Android-MCP-go/releases/latest/download/android-mcp-windows-$Arch.exe"
    $ZipUrl     = "https://github.com/tintupratap/Android-MCP-go/releases/latest/download/android-mcp-windows-$Arch.zip"
}

try {
    Invoke-WebRequest -Uri $ReleaseUrl -OutFile $BinaryTarget -UseBasicParsing -ErrorAction Stop
    $Installed = $true
    Write-HostColor "✓ Downloaded prebuilt release binary from GitHub Releases!" "Green"
} catch {
    Write-HostColor "Direct release binary download failed. Attempting release zip extraction..." "Yellow"
    $TempZip = Join-Path $env:TEMP "android-mcp-release.zip"
    $TempExtract = Join-Path $env:TEMP "android-mcp-extract"

    try {
        Invoke-WebRequest -Uri $ZipUrl -OutFile $TempZip -UseBasicParsing -ErrorAction Stop
        Expand-Archive -Path $TempZip -DestinationPath $TempExtract -Force
        $ExtractedBin = Get-ChildItem -Path $TempExtract -Recurse -Filter "android-mcp.exe" | Select-Object -First 1
        if ($ExtractedBin) {
            Copy-Item -Path $ExtractedBin.FullName -Destination $BinaryTarget -Force
            $Installed = $true
            Write-HostColor "✓ Extracted prebuilt release binary from zip archive!" "Green"
        }
    } catch {} finally {
        Remove-Item -Path $TempZip -ErrorAction SilentlyContinue
        Remove-Item -Path $TempExtract -Recurse -ErrorAction SilentlyContinue
    }
}

# 5. Fallback: Build from Source if Go Compiler is Installed
if (-not $Installed) {
    if (Get-Command go -ErrorAction SilentlyContinue) {
        Write-HostColor "Release download unavailable. Falling back to building from source using Go..." "Yellow"
        $TempRepo = Join-Path $env:TEMP "android-mcp-repo"
        Remove-Item -Path $TempRepo -Recurse -ErrorAction SilentlyContinue
        
        try {
            git clone --depth 1 https://github.com/tintupratap/Android-MCP-go.git $TempRepo
            Push-Location $TempRepo
            go build -ldflags="-s -w -X main.Version=0.5.0" -o $BinaryTarget ./cmd/android-mcp
            Pop-Location
            $Installed = $true
            Write-HostColor "✓ Built and installed binary from source!" "Green"
        } catch {} finally {
            Remove-Item -Path $TempRepo -Recurse -ErrorAction SilentlyContinue
        }
    }
}

if (-not $Installed) {
    Write-HostColor "❌ Failed to install android-mcp. Check network connection or releases page: https://github.com/tintupratap/Android-MCP-go/releases" "Red"
    exit 1
}

Write-HostColor "✓ Binary installed to $BinaryTarget" "Green"

# 6. Ensure Platform-Tools, scrcpy Display Mirror & Skills
Write-HostColor "Ensuring official Android Platform-Tools..." "Cyan"
& "$BinaryTarget" platform-tools update 2>&1 | Out-Null

Write-HostColor "Ensuring official scrcpy display mirror..." "Cyan"
& "$BinaryTarget" scrcpy update 2>&1 | Out-Null

Write-HostColor "Ensuring machine-readable skills manifests..." "Cyan"
& "$BinaryTarget" skills install 2>&1 | Out-Null

# 7. Health Check
Write-HostColor "Running installation health check..." "Cyan"
& "$BinaryTarget" doctor

Write-HostColor "✓ Android-MCP-go installation complete and verified!" "Green"
Write-HostColor "Releases page: https://github.com/tintupratap/Android-MCP-go/releases" "Cyan"
