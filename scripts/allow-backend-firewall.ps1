# Windows güvenlik duvarında backend portunu (TCP 8080) telefona açar.
# Yönetici olarak çalıştırın:
#   powershell -ExecutionPolicy Bypass -File scripts/allow-backend-firewall.ps1
param([int]$Port = 8080)

$ErrorActionPreference = "Stop"

$identity = [Security.Principal.WindowsIdentity]::GetCurrent()
$principal = New-Object Security.Principal.WindowsPrincipal($identity)
if (-not $principal.IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)) {
    Write-Warning "Bu betik yönetici yetkisi ister. PowerShell'i 'Yönetici olarak çalıştır' ile açın."
    exit 1
}

$ruleName = "AVATAR backend TCP $Port"
$existing = Get-NetFirewallRule -DisplayName $ruleName -ErrorAction SilentlyContinue
if ($existing) {
    Write-Output "Kural zaten var: $ruleName"
} else {
    New-NetFirewallRule `
        -DisplayName $ruleName `
        -Direction Inbound `
        -Action Allow `
        -Protocol TCP `
        -LocalPort $Port `
        -Profile Private, Domain | Out-Null
    Write-Output "Kural eklendi: $ruleName"
}

$lanIps = Get-NetIPAddress -AddressFamily IPv4 |
    Where-Object { $_.IPAddress -notlike "127.*" -and $_.IPAddress -notlike "169.254.*" } |
    Select-Object -ExpandProperty IPAddress
Write-Output "PC LAN IP adayları:"
$lanIps | ForEach-Object { Write-Output "  http://$($_):$Port" }
Write-Output "Telefonun bu adreslerden birine erişebildiğini doğrulayın."
