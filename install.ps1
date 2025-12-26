$ErrorActionPreference = "Stop"

$repo = "WangShayne/cgpd"
$binaryName = "cgpd"

$arch = if ([Environment]::Is64BitOperatingSystem) { "amd64" } else { "386" }
$filename = "${binaryName}-windows-${arch}.exe"
$downloadUrl = "https://github.com/${repo}/releases/latest/download/${filename}"

$installDir = "$env:LOCALAPPDATA\Programs\cgpd"
$installPath = Join-Path $installDir "${binaryName}.exe"

Write-Host "Detected: windows/${arch}"
Write-Host "Downloading ${filename}..."

if (-not (Test-Path $installDir)) {
    New-Item -ItemType Directory -Path $installDir -Force | Out-Null
}

Invoke-WebRequest -Uri $downloadUrl -OutFile $installPath -UseBasicParsing

$userPath = [Environment]::GetEnvironmentVariable("Path", "User")
if ($userPath -notlike "*$installDir*") {
    [Environment]::SetEnvironmentVariable("Path", "$userPath;$installDir", "User")
    $env:Path = "$env:Path;$installDir"
    Write-Host "Added $installDir to PATH"
}

Write-Host "Installed: $installPath"
& $installPath --version
