$env:WSL_UTF8 = 1
$OutputEncoding = [System.Text.Encoding]::UTF8
$currentPath = Get-Location
$architecture = $env:PROCESSOR_ARCHITECTURE
$downloadCdnUrlFromEnv = $env:DOWNLOAD_CDN_URL
$version = "#__VERSION__"
$downloadUrl = "https://dc3p1870nn3cj.cloudfront.net"

function Test-Wait {
  while ($true) {
    Start-Sleep -Seconds 1
  }
}

$runAsAdmin = New-Object Security.Principal.WindowsPrincipal([Security.Principal.WindowsIdentity]::GetCurrent())
if (-not $runAsAdmin.IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)) {
    Write-Host "`n`nThe installation script needs to be run as an administrator.`n"
    Write-Host "Please try the following methods:`n"
    Write-Host "1. Search for 'PowerShell' in the Start menu, right-click it, and select 'Run as administrator'. "
    Write-Host "   Navigate to the directory where the installation script is located and run the installation script.`n"
    Write-Host "2. Press Win + R, type 'powershell', and then press Ctrl + Shift + Enter. "
    Write-Host "   Navigate to the directory where the installation script is located and run the installation script.`n"
    Write-Host "`nPress Ctrl+C to exit.`n"
    Test-Wait
}

$process = Get-Process -Name olares-cli -ErrorAction SilentlyContinue
if ($process) {
  Write-Host "olares-cli.exe is running, Press Ctrl+C to exit."
  Test-Wait
}

$distro = wsl --list | Select-String -Pattern "^Ubuntu$"
if (-not $distro -eq "") {
  Write-Host "Distro Olares exists, please unregister it first."
  exit 1
}

$arch = "amd64"
if ($architecture -like "ARM") {
  $arch = "arm64"
}

if (-Not $downloadCdnUrlFromEnv -eq "") {
  $downloadUrl = $downloadCdnUrlFromEnv
}

$CLI_PROGRAM_PATH = "{0}\" -f $currentPath
if (-Not (Test-Path $CLI_PROGRAM_PATH)) {
  New-Item -Path $CLI_PROGRAM_PATH -ItemType Directory
}

$CLI_VERSION = "0.2.5"
$CLI_FILE = "olares-cli-v{0}_windows_{1}.tar.gz" -f $CLI_VERSION, $arch
$CLI_URL = "{0}/{1}" -f $downloadUrl, $CLI_FILE
$CLI_PATH = "{0}{1}" -f $CLI_PROGRAM_PATH, $CLI_FILE

$download = 0
if (Test-Path $CLI_PATH) {
  tar -xzf $CLI_PATH -C $CLI_PROGRAM_PATH *> $null
  if (-Not ($LASTEXITCODE -eq 0)) {
    Remove-Item -Path $CLI_PATH
    $download = 1
  }
} else {
  $download = 1
}

if ($download -eq 1) {
  curl -Uri $CLI_URL -OutFile $CLI_PATH
  Write-Host "Downloading olares-cli.exe..."
  if (-Not (Test-Path $CLI_PATH)) {
    Write-Host "Download olares-cli.exe failed."
    exit 1
  }
  tar -xzf $CLI_PATH -C $CLI_PROGRAM_PATH *> $null
  $cliPath = "{0}\olares-cli.exe" -f $CLI_PROGRAM_PATH
  if ( -Not (Test-Path $cliPath)) {
    Write-Host "olares-cli.exe not found."
    exit 1
  }
}

Start-Sleep -Seconds 3
Write-Host ("Preparing to start the installation of Olares {0}. Depending on your network conditions, this process may take several minutes." -f $version)

$command = "{0}\olares-cli.exe olares install --version {1}" -f $CLI_PROGRAM_PATH, $version
Start-Process cmd -ArgumentList '/k',$command -Wait -Verb RunAs

