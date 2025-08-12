param([string]$Tag = "latest")

# Set $env:REPO="YOURUSER/opencode-gpt5-fork" before running or edit below.
$repo = if ($env:REPO) { $env:REPO } else { "YOURUSER/opencode-gpt5-fork" }
$arch = "amd64"
$os = "windows"

if ($Tag -eq "latest") {
  $url = "https://github.com/$repo/releases/latest/download/ocx_${os}_${arch}.zip"
} else {
  $url = "https://github.com/$repo/releases/download/$Tag/ocx_${os}_${arch}.zip"
}

$temp = New-Item -ItemType Directory -Force -Path ([IO.Path]::GetTempPath() + [IO.Path]::GetRandomFileName())
$zip = Join-Path $temp "ocx.zip"
Invoke-WebRequest -Uri $url -OutFile $zip -UseBasicParsing
Expand-Archive $zip -DestinationPath $temp -Force

$dest = "$Env:ProgramFiles\ocx"
New-Item -ItemType Directory -Force -Path $dest | Out-Null
Copy-Item "$temp\ocx.exe" $dest -Force

$envPath = [Environment]::GetEnvironmentVariable("Path", "Machine")
if ($envPath -notlike "*$dest*") {
  [Environment]::SetEnvironmentVariable("Path", "$envPath;$dest", "Machine")
  Write-Host "Added $dest to PATH (system). You may need a new terminal."
}

Write-Host "Installed ocx -> $dest\ocx.exe"
