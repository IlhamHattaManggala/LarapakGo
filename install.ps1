# ==============================================================================
# LarapakGo Windows One-Line Installer (PowerShell)
# Dialek: Ngapak Tegalan Murni
# Cara nganggo:
#   irm https://raw.githubusercontent.com/IlhamHattaManggala/LarapakGo/main/install.ps1 | iex
# ==============================================================================

$ErrorActionPreference = "Stop"

$AppName = "LarapakGo"
$Version = "1.2.2"
$RepoOwner = "IlhamHattaManggala"
$RepoName = "LarapakGo"
$Branch = "main"

Write-Host ""
Write-Host "==========================================================" -ForegroundColor DarkGray
Write-Host "   LarapakGo CLI Installer - Versi $Version" -ForegroundColor Yellow
Write-Host "   Framework Web MVC Basa Go & Interpreter NgapakIn (.ngpk)" -ForegroundColor Cyan
Write-Host "==========================================================" -ForegroundColor DarkGray
Write-Host ""

# 1. Nemtokna Arsitektur Sistem
$Arch = if ([IntPtr]::Size -eq 8) { "amd64" } else { "386" }
if ($Arch -ne "amd64") {
    Write-Host "[Kliru] Saiki LarapakGo mung ndukung Windows 64-bit (amd64)." -ForegroundColor Red
    exit 1
}

$BinaryName = "larapak-v$Version-windows-amd64.exe"
$DownloadUrl = "https://raw.githubusercontent.com/$RepoOwner/$RepoName/$Branch/dist/$BinaryName"

# 2. Nemtokna Folder Instalasi
$InstallDir = Join-Path $env:USERPROFILE ".larapak\bin"
if (-not (Test-Path $InstallDir)) {
    New-Item -ItemType Directory -Path $InstallDir -Force | Out-Null
}

$TargetExe = Join-Path $InstallDir "larapak.exe"

# 3. Ngundhuh Binary
Write-Host "[+] Ngundhuh LarapakGo saka GitHub..." -ForegroundColor Cyan
try {
    [Net.ServicePointManager]::SecurityProtocol = [Net.SecurityProtocolType]::Tls12
    Invoke-WebRequest -Uri $DownloadUrl -OutFile $TargetExe -UseBasicParsing
    Write-Host "[OK] Berkas binary sukses diundhuh maring: $TargetExe" -ForegroundColor Green
} catch {
    Write-Host "[Pèngetan] Gagal ngundhuh saka dist, nyoba ngundhuh binary utama..." -ForegroundColor Yellow
    $FallbackUrl = "https://raw.githubusercontent.com/$RepoOwner/$RepoName/$Branch/larapak.exe"
    try {
        Invoke-WebRequest -Uri $FallbackUrl -OutFile $TargetExe -UseBasicParsing
        Write-Host "[OK] Berkas binary sukses diundhuh saka fallback." -ForegroundColor Green
    } catch {
        Write-Host "[Kliru] Gagal ngundhuh binary LarapakGo: $_" -ForegroundColor Red
        exit 1
    }
}

# 4. Nambahna maring Environment Variables (PATH)
$UserPath = [Environment]::GetEnvironmentVariable("Path", "User")
if ($UserPath -notlike "*$InstallDir*") {
    Write-Host "[+] Nambahna '$InstallDir' maring PATH User..." -ForegroundColor Cyan
    $NewPath = if ([string]::IsNullOrEmpty($UserPath)) { $InstallDir } else { "$UserPath;$InstallDir" }
    [Environment]::SetEnvironmentVariable("Path", $NewPath, "User")
    Write-Host "[OK] PATH User sukses dianyari." -ForegroundColor Green
} else {
    Write-Host "[OK] '$InstallDir' wis ana neng PATH User." -ForegroundColor DarkGray
}

# Nyelaraskna PATH neng sesi terminal saiki
if ($env:PATH -notlike "*$InstallDir*") {
    $env:PATH = "$InstallDir;$env:PATH"
}

# 5. Priksa Instalasi
Write-Host ""
if (Test-Path $TargetExe) {
    Write-Host "==========================================================" -ForegroundColor Green
    Write-Host " [SUKSES] LarapakGo v$Version sukses kepasang neng komputermu!" -ForegroundColor Green
    Write-Host "==========================================================" -ForegroundColor Green
    Write-Host ""
    Write-Host "Lokasi eksekusi: $TargetExe" -ForegroundColor DarkGray
    Write-Host ""
    Write-Host "Langkah sateruse nggo miwiti:" -ForegroundColor Yellow
    Write-Host "  1. Bukak terminal anyar (PowerShell / Command Prompt)" -ForegroundColor White
    Write-Host "  2. Tes perintah CLI:" -ForegroundColor White
    Write-Host "       larapak tulung" -ForegroundColor Cyan
    Write-Host "  3. Gawe proyek web anyar nganggo dialek Ngapak:" -ForegroundColor White
    Write-Host "       larapak anyar warung_sate --bahasa=ngapak" -ForegroundColor Cyan
    Write-Host "  4. Mlaku maring folder lan ladeni server:" -ForegroundColor White
    Write-Host "       cd warung_sate" -ForegroundColor Cyan
    Write-Host "       larapak ladeni" -ForegroundColor Cyan
    Write-Host ""
    Write-Host "Ora Ngapak Ora Kepenak!" -ForegroundColor Yellow
    Write-Host ""
} else {
    Write-Host "[Kliru] Berkas eksekusi ora ditemokna sawise instalasi." -ForegroundColor Red
    exit 1
}

