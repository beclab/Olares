$currentPath = Get-Location
$architecture = $env:PROCESSOR_ARCHITECTURE
$version = "#__VERSION__"

function Test-IsAdmin {
  $currentUser = New-Object Security.Principal.WindowsPrincipal $([Security.Principal.WindowsIdentity]::GetCurrent())
  $currentUser.IsInRole("Administrators")
}

if (-Not (Test-IsAdmin)) {
  Write-Host "Not running as Administrator"
  exit 1
}

$process = Get-Process -Name olares-cli -ErrorAction SilentlyContinue
if ($process) {
  Write-Host "olares-cli.exe is running, exit."
  exit 1
}

$arch = "amd64"
if ($architecture -like "ARM") {
  $arch = "arm64"
}

$CLI_VERSION = "0.1.47"
$CLI_FILE = "olares-cli-v{0}_windows_{1}.tar.gz" -f $CLI_VERSION, $arch
$CLI_URL = "https://dc3p1870nn3cj.cloudfront.net/{0}" -f $CLI_FILE
$CLI_PATH = "{0}\{1}"  -f $currentPath, $CLI_FILE
if (-Not (Test-Path $CLI_FILE)) {
  curl -Uri $CLI_URL -OutFile $CLI_PATH
}

if (-Not (Test-Path $CLI_PATH)) {
  Write-Host "Download olares-cli.exe failed."
  exit 1
}

tar -xf $CLI_PATH
$cliPath = "{0}\olares-cli.exe" -f $currentPath
if ( -Not (Test-Path $cliPath)) {
  Write-Host "olares-cli.exe not found."
  exit 1
}

wsl --unregister Ubuntu *> $null

Start-Sleep -Seconds 3
$arguments = @("terminus", "install", "--version", $version)
Write-Host ("Preparing to start the installation of Olares {0}. Depending on your network conditions, this process may take several minutes." -f $version)
Start-Process -FilePath $cliPath -ArgumentList $arguments -Wait
