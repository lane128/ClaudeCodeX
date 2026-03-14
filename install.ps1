$ErrorActionPreference = "Stop"

$Repo   = "lane128/ClaudeCodeX"
$Binary = "ccx"

# --- detect arch ---
$Arch = if ($env:PROCESSOR_ARCHITECTURE -eq "ARM64") { "arm64" } else { "amd64" }

# --- download ---
Write-Host "Downloading $Binary (windows/$Arch)..."
$Url = "https://github.com/$Repo/releases/latest/download/${Binary}_windows_${Arch}.exe"
$InstallDir = "$env:USERPROFILE\.local\bin"
New-Item -ItemType Directory -Force -Path $InstallDir | Out-Null
$InstallPath = "$InstallDir\$Binary.exe"
Invoke-WebRequest -Uri $Url -OutFile $InstallPath

# --- add to user PATH if missing ---
$UserPath = [Environment]::GetEnvironmentVariable("PATH", "User")
if ($UserPath -notlike "*$InstallDir*") {
  [Environment]::SetEnvironmentVariable("PATH", "$UserPath;$InstallDir", "User")
  Write-Host "Added $InstallDir to PATH (effective in new shells)"
}

Write-Host "Installed: $InstallPath"

# --- initialize default settings ---
& $InstallPath setting | Out-Null
Write-Host "Default config: $env:USERPROFILE\.ccx\settings.json"

Write-Host ""
Write-Host "Run 'ccx doctor' to check your network connectivity."
