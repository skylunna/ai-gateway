# run.ps1 - One-command runner for luner Python SDK example (Windows)
# Supports: PowerShell 5.1+ / PowerShell 7+

param(
    [switch]$Help
)

$ScriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
Set-Location $ScriptDir

# Colors (Clean, No Emojis)
$Green = "Green"
$Yellow = "Yellow"
$Red = "Red"
$Default = "White"

function Write-Info    { param($msg) Write-Host "[INFO] $msg" -ForegroundColor $Green }
function Write-Warn    { param($msg) Write-Host "[WARN] $msg" -ForegroundColor $Yellow }
function Write-Error   { param($msg) Write-Host "[ERROR] $msg" -ForegroundColor $Red }

# Check dependencies
function Check-Deps {
    if (-not (Get-Command uv -ErrorAction SilentlyContinue)) {
        Write-Error "uv not found. Please install: https://docs.astral.sh/uv/getting-started/installation/"
        exit 1
    }
}

# Check gateway health
function Check-Gateway {
    Write-Info "Checking luner gateway health..."
    try {
        $response = Invoke-WebRequest -Uri "http://localhost:8080/health" -UseBasicParsing -TimeoutSec 5
        if ($response.StatusCode -ne 200) { throw "Status $($response.StatusCode)" }
    } catch {
        Write-Warn "Gateway not responding at http://localhost:8080"
        Write-Host ""
        Write-Host "Please start luner first:"
        Write-Host "  go run cmd/aigw/main.go -config config.yaml"
        Write-Host ""
    }
}

# Run tests
function Run-Tests {
    Write-Info "Running integration tests with uv..."
    uv run python test/test_integration.py
}

# Main
function Main {
    if ($Help) {
        Write-Host "Usage: .\run.ps1 [OPTIONS]"
        Write-Host ""
        Write-Host "One-command runner for luner Python SDK example."
        Write-Host ""
        Write-Host "Options:"
        Write-Host "  -Help     Show this help message"
        exit 0
    }

    Write-Info "Starting luner Python SDK example runner"
    Check-Deps
    Check-Gateway
    Run-Tests
    Write-Info "All done!"
}

Main