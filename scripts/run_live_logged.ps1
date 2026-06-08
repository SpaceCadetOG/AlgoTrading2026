$goLive = Read-Host "Go live? [y/N]"
if ($goLive -match '^[Yy]$') {
    $env:TRADING_MODE = 'live'
} else {
    $env:TRADING_MODE = 'paper'
}

New-Item -ItemType Directory -Force -Path .\data\logs | Out-Null
$logFile = ".\data\logs\runtime_$((Get-Date).ToUniversalTime().ToString('yyyyMMddTHHmmssZ')).log"

go run ./cmd/main.go 2>&1 | Tee-Object -FilePath $logFile
