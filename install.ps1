# -----------------------------------------------------------
# Copyright (c) Will Velida. All rights reserved.
# Licensed under the MIT License.
# -----------------------------------------------------------

param (
    [string]$Version,
    [string]$InstallDir = "$Env:SystemDrive\code-minions"
)

Write-Output ""
$ErrorActionPreference = 'stop'

# Escape spaces in path
$InstallDir = $InstallDir -replace ' ', '` '

# Constants
$BinaryName = "code-minions.exe"
$BinaryPath = "${InstallDir}\${BinaryName}"
$GitHubOrg = "willvelida"
$GitHubRepo = "code-minions"

# Detect architecture
$Arch = "amd64"
if ($env:PROCESSOR_ARCHITECTURE -eq "ARM64" -or $env:PROCESSOR_IDENTIFIER -like "*ARM*") {
    $Arch = "arm64"
}
Write-Output "Detected architecture: $Arch"

# Verify execution policy allows running scripts
$policy = Get-ExecutionPolicy
if ($policy -eq 'Restricted' -or $policy -eq 'AllSigned') {
    Write-Output "PowerShell execution policy '$policy' does not allow running this script."
    Write-Output "To make this change please run:"
    Write-Output "  Set-ExecutionPolicy RemoteSigned -Scope CurrentUser"
    break
}

# Support TLS 1.2 for older PowerShell versions
[Net.ServicePointManager]::SecurityProtocol = "tls12, tls11, tls"

# Set up GitHub authentication if available
if ($Env:GITHUB_USER) {
    $basicAuth = [System.Convert]::ToBase64String(
        [System.Text.Encoding]::ASCII.GetBytes($Env:GITHUB_USER + ":" + $Env:GITHUB_TOKEN)
    )
    $githubHeader = @{ "Authorization" = "Basic $basicAuth" }
} else {
    $githubHeader = @{}
}

# Check for existing installation
if (Test-Path $BinaryPath -PathType Leaf) {
    Write-Warning "code-minions is detected - $BinaryPath"
    & $BinaryPath version
    Write-Output "Reinstalling code-minions..."
} else {
    Write-Output "Installing code-minions..."
}

# Create install directory
Write-Output "Creating $InstallDir directory"
New-Item -ErrorAction Ignore -Path $InstallDir -ItemType "directory" | Out-Null
if (!(Test-Path $InstallDir -PathType Container)) {
    Write-Warning "Could not create $InstallDir. You may need to run as Administrator,"
    Write-Warning "or specify a custom install directory:"
    Write-Warning "  .\install.ps1 -InstallDir `"`$Env:LOCALAPPDATA\Programs\code-minions`""
    throw "Cannot create $InstallDir"
}

# Fetch releases from GitHub
$releaseUrl = "https://api.github.com/repos/${GitHubOrg}/${GitHubRepo}/releases"
$releases = Invoke-RestMethod -Headers $githubHeader -Uri $releaseUrl -Method Get
if ($releases.Count -eq 0) {
    throw "No releases found at github.com/${GitHubOrg}/${GitHubRepo}"
}

# Resolve version
if (!$Version) {
    $release = $releases | Where-Object { $_.tag_name -notlike "*rc*" } | Select-Object -First 1
} else {
    $release = $releases | Where-Object { $_.tag_name -eq "v$Version" } | Select-Object -First 1
}

if (!$release) {
    throw "Cannot find the specified version"
}

Write-Output "Installing code-minions $($release.tag_name)..."

# Find the zip asset for the detected architecture
$assetPattern = "*windows_${Arch}.zip"
$asset = $release | Select-Object -ExpandProperty assets | Where-Object { $_.name -like $assetPattern }
if (!$asset) {
    throw "Cannot find a Windows ${Arch} binary for $($release.tag_name)"
}

# Download the archive to a temp directory for clean failure handling
$tempDir = Join-Path ([System.IO.Path]::GetTempPath()) "code-minions-install"
New-Item -ErrorAction Ignore -Path $tempDir -ItemType "directory" | Out-Null

$zipFileName = $asset.name
$zipFilePath = Join-Path $tempDir $zipFileName

try {

Write-Output "Downloading $($asset.browser_download_url)..."
$oldProgressPreference = $ProgressPreference
$ProgressPreference = 'SilentlyContinue'
Invoke-WebRequest -Uri $asset.browser_download_url -OutFile $zipFilePath
$ProgressPreference = $oldProgressPreference

if (!(Test-Path $zipFilePath -PathType Leaf)) {
    throw "Failed to download code-minions - $zipFilePath"
}

# Verify checksum
$checksumAsset = $release | Select-Object -ExpandProperty assets | Where-Object { $_.name -eq "checksums.txt" }
if ($checksumAsset) {
    Write-Output "Verifying checksum..."
    $checksumUrl = $checksumAsset.browser_download_url
    $oldProgressPreference2 = $ProgressPreference
    $ProgressPreference = 'SilentlyContinue'
    $checksumContent = (Invoke-WebRequest -Uri $checksumUrl).Content
    $ProgressPreference = $oldProgressPreference2
    $expectedHash = ($checksumContent -split "`n" | Where-Object { $_ -like "*$zipFileName*" } | ForEach-Object { ($_ -split "\s+")[0] }).Trim()
    $actualHash = (Get-FileHash -Path $zipFilePath -Algorithm SHA256).Hash.ToLower()
    if ($expectedHash -and $actualHash -ne $expectedHash) {
        Remove-Item $zipFilePath -Force
        throw "Checksum verification failed. Expected: $expectedHash, Got: $actualHash"
    }
    Write-Output "Checksum verified."
} else {
    Write-Warning "Checksum file not found in release - skipping verification"
}

# Extract the archive
Write-Output "Extracting $zipFilePath..."
Microsoft.PowerShell.Archive\Expand-Archive -Force -Path $zipFilePath -DestinationPath $InstallDir
if (!(Test-Path $BinaryPath -PathType Leaf)) {
    throw "Failed to extract code-minions binary"
}

# Verify installation
& $BinaryPath version

} finally {
    # Clean up temp directory
    if (Test-Path $tempDir) {
        Write-Output "Cleaning up..."
        Remove-Item $tempDir -Recurse -Force
    }
}

# Add to User PATH if not already present
Write-Output "Checking PATH..."
$userPath = [Environment]::GetEnvironmentVariable("PATH", "User")
if ($userPath -like "*code-minions*") {
    Write-Output "PATH already contains code-minions directory - skipping"
} else {
    [Environment]::SetEnvironmentVariable("PATH", $userPath + ";$InstallDir", "User")
    Write-Output "Added $InstallDir to User PATH"
    Write-Output "Restart your terminal for the PATH change to take effect."
}

Write-Output ""
Write-Output "code-minions CLI installed successfully to $BinaryPath"
Write-Output "Run 'code-minions --help' to get started."
