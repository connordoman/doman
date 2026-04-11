# PowerShell installation script for doman
# Usage: .\install.ps1 or: irm https://raw.githubusercontent.com/connordoman/doman/main/install.ps1 | iex

$ErrorActionPreference = "Stop"

$VERSION = if ($env:VERSION) { $env:VERSION } else { "latest" }
$REPO = "connordoman/doman"

# Determine install directory (prefer ~\.local\bin to avoid admin)
if (-not $env:INSTALL_DIR) {
    $LocalBinDir = Join-Path $env:USERPROFILE ".local\bin"
    try {
        if (-not (Test-Path $LocalBinDir)) {
            New-Item -ItemType Directory -Path $LocalBinDir -Force | Out-Null
        }
        $INSTALL_DIR = $LocalBinDir
    } catch {
        $INSTALL_DIR = Join-Path $env:ProgramFiles "doman"
    }
} else {
    $INSTALL_DIR = $env:INSTALL_DIR
}

# Detect architecture
$Arch = if ([Environment]::Is64BitOperatingSystem) { "x86_64" } else { "i386" }
if ($env:PROCESSOR_ARCHITECTURE -eq "ARM64" -or $env:PROCESSOR_ARCHITEW6432 -eq "ARM64") {
    $Arch = "arm64"
}

$OS_NAME = "Windows"
$FILE_NAME = "doman_${OS_NAME}_${Arch}.zip"

# Fetch latest version if needed
if ($VERSION -eq "latest") {
    Write-Host "Fetching latest version..."
    try {
        $response = Invoke-RestMethod -Uri "https://api.github.com/repos/$REPO/releases/latest" -ErrorAction Stop
        $VERSION = $response.tag_name -replace '^v', ''
    } catch {
        Write-Error "Failed to fetch latest version from GitHub API"
        exit 1
    }
} else {
    $VERSION = $VERSION -replace '^v', ''
}

$DOWNLOAD_URL = "https://github.com/${REPO}/releases/download/v${VERSION}/${FILE_NAME}"

# Check if already installed
if (Get-Command doman -ErrorAction SilentlyContinue) {
    try {
        $CurrentVersion = (doman version 2>$null) -replace '.*v?(\d+\.\d+\.\d+).*', '$1'
        Write-Host "doman is already installed (version: $CurrentVersion)"
    } catch {
        Write-Host "doman is already installed (version: unknown)"
    }
    Write-Host "Installing version v${VERSION}..."
} else {
    Write-Host "Installing doman v${VERSION}..."
}

# Create temporary directory
$TMP_DIR = New-TemporaryFile | ForEach-Object { Remove-Item $_; New-Item -ItemType Directory -Path $_ }

try {
    # Download release archive
    Write-Host "Downloading from $DOWNLOAD_URL..."
    $ZipPath = Join-Path $TMP_DIR $FILE_NAME
    try {
        Invoke-WebRequest -Uri $DOWNLOAD_URL -OutFile $ZipPath -ErrorAction Stop
    } catch {
        Write-Error "Failed to download release archive`nURL: $DOWNLOAD_URL"
        exit 1
    }

    # Extract archive
    Write-Host "Extracting archive..."
    try {
        Expand-Archive -Path $ZipPath -DestinationPath $TMP_DIR -Force
    } catch {
        Write-Error "Failed to extract archive"
        exit 1
    }

    # Find binary (Windows binary has .exe extension)
    $Binary = Get-ChildItem -Path $TMP_DIR -Filter "doman.exe" -Recurse -File | Select-Object -First 1
    if (-not $Binary) {
        Write-Error "Could not find doman.exe binary in archive"
        exit 1
    }

    # Ensure install directory exists
    if (-not (Test-Path $INSTALL_DIR)) {
        Write-Host "Creating directory $INSTALL_DIR..."
        try {
            New-Item -ItemType Directory -Path $INSTALL_DIR -Force | Out-Null
        } catch {
            Write-Error "Failed to create directory $INSTALL_DIR. You may need to run as administrator."
            exit 1
        }
    }

    # Install binary
    $TargetPath = Join-Path $INSTALL_DIR "doman.exe"
    Write-Host "Installing to $INSTALL_DIR..."
    Copy-Item -Path $Binary.FullName -Destination $TargetPath -Force

    Write-Host "✓ Successfully installed doman v${VERSION}"

    # Verify installation
    if (Get-Command doman -ErrorAction SilentlyContinue) {
        doman version
    } else {
        Write-Warning "doman was installed but is not in PATH"
        Write-Warning "You may need to add $INSTALL_DIR to your PATH"
        if ($INSTALL_DIR -eq $LocalBinDir) {
            Write-Host "Run: `$env:Path += `";$INSTALL_DIR`"" -ForegroundColor Yellow
            Write-Host "Or add it permanently: [Environment]::SetEnvironmentVariable('Path', [Environment]::GetEnvironmentVariable('Path', 'User') + `";$INSTALL_DIR`", 'User')" -ForegroundColor Yellow
        }
        & $TargetPath version
    }
} finally {
    # Cleanup
    Remove-Item -Path $TMP_DIR -Recurse -Force -ErrorAction SilentlyContinue
}
