# scripts/deploy_local.ps1
$ErrorActionPreference = "Stop"

# Configuration
$Registry = "192.168.220.128:5000"
$Tag = "v1.1.0"
$Services = @("api", "booking-service", "bus-service", "payment-service", "notification-service", "user-service", "ws-service")

# 1. Locate OpenSSH path resolving Windows File System Redirector (Sysnative vs System32)
$OpenSSHPath = ""
if (Test-Path "C:\Windows\System32\OpenSSH\scp.exe") {
    $OpenSSHPath = "C:\Windows\System32\OpenSSH"
} elseif (Test-Path "C:\Windows\Sysnative\OpenSSH\scp.exe") {
    $OpenSSHPath = "C:\Windows\Sysnative\OpenSSH"
}

if ($OpenSSHPath) {
    $env:Path = "$OpenSSHPath;" + $env:Path
}

# 2. Ensure env variables for cross compilation
$env:GOOS = "linux"
$env:GOARCH = "amd64"
$env:CGO_ENABLED = "0"

Write-Host "=== Starting local build & deploy process ===" -ForegroundColor Green

# Create temp build directory
New-Item -ItemType Directory -Force -Path "bin" | Out-Null

# 2. Build & Push each service
foreach ($Service in $Services) {
    Write-Host "Building service: ${Service}..." -ForegroundColor Cyan
    
    # Compile binary
    go build -ldflags="-w -s" -o "bin/service" "cmd/${Service}/main.go"
    
    Write-Host "Docker building and pushing: ${Service}..." -ForegroundColor Cyan
    
    $ImageName = "${Registry}/bus-booking-${Service}:${Tag}"
    
    # Docker Build
    docker build -t $ImageName -f Dockerfile.local .
    
    # Docker Push
    docker push $ImageName
}

# Cleanup temp files
Remove-Item -Path "bin/service" -Force

# 3. SCP Stack configuration to manager VM
Write-Host "Uploading app-stack.yml to manager node..." -ForegroundColor Cyan
scp -o StrictHostKeyChecking=no -i deployment/local/id_rsa deployment/local/app-stack.yml manager@192.168.220.128:/home/manager/app-stack.yml

# 4. Deploy stack on Swarm Manager
Write-Host "Deploying stack bus-booking on Swarm..." -ForegroundColor Green
ssh -o StrictHostKeyChecking=no -i deployment/local/id_rsa manager@192.168.220.128 "DOCKER_REGISTRY=$Registry TAG=$Tag docker stack deploy -c /home/manager/app-stack.yml bus-booking"

Write-Host "=== Deployment Completed Successfully ===" -ForegroundColor Green
