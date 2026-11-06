#!/usr/bin/env bash
# ==============================================================================
# LarapakGo Linux & macOS One-Line Installer
# Dialek: Ngapak Tegalan Murni
# Cara nganggo:
#   curl -fsSL https://raw.githubusercontent.com/IlhamHattaManggala/LarapakGo/main/install.sh | bash
# ==============================================================================

set -e

APP_NAME="LarapakGo"
VERSION="1.2.2"
REPO_OWNER="IlhamHattaManggala"
REPO_NAME="LarapakGo"
BRANCH="main"

echo ""
echo "=========================================================="
echo "   LarapakGo CLI Installer - Versi $VERSION"
echo "   Framework Web MVC Basa Go & Interpreter NgapakIn (.ngpk)"
echo "=========================================================="
echo ""

# 1. Nemtokna Sistem Operasi & Arsitektur
OS="$(uname -s)"
ARCH="$(uname -m)"

BINARY_NAME=""

case "$OS" in
    Linux)
        case "$ARCH" in
            x86_64|amd64)
                BINARY_NAME="larapak-v$VERSION-linux-amd64"
                ;;
            *)
                echo "[Kliru] Arsitektur Linux $ARCH durung didukung sacara langsung."
                exit 1
                ;;
        esac
        ;;
    Darwin)
        case "$ARCH" in
            arm64|aarch64)
                BINARY_NAME="larapak-v$VERSION-darwin-arm64"
                ;;
            x86_64|amd64)
                BINARY_NAME="larapak-v$VERSION-darwin-arm64"
                ;;
            *)
                echo "[Kliru] Arsitektur macOS $ARCH durung didukung sacara langsung."
                exit 1
                ;;
        esac
        ;;
    *)
        echo "[Kliru] Sistem operasi $OS ora didukung dening installer kiye."
        exit 1
        ;;
esac

DOWNLOAD_URL="https://raw.githubusercontent.com/$REPO_OWNER/$REPO_NAME/$BRANCH/dist/$BINARY_NAME"

# 2. Nemtokna Folder Instalasi
INSTALL_DIR="$HOME/.larapak/bin"
mkdir -p "$INSTALL_DIR"
TARGET_BIN="$INSTALL_DIR/larapak"

# 3. Ngundhuh Binary
echo "[+] Ngundhuh $BINARY_NAME saka GitHub..."
if command -v curl >/dev/null 2>&1; then
    curl -fsSL "$DOWNLOAD_URL" -o "$TARGET_BIN"
elif command -v wget >/dev/null 2>&1; then
    wget -qO "$TARGET_BIN" "$DOWNLOAD_URL"
else
    echo "[Kliru] Butuh curl utawa wget nggo ngundhuh berkas."
    exit 1
fi

chmod +x "$TARGET_BIN"
echo "[OK] Berkas binary sukses diundhuh maring: $TARGET_BIN"

# 4. Nambahna maring PATH Profil Shell
SHELL_CONFIG=""
if [ -n "$ZSH_VERSION" ] || [ -f "$HOME/.zshrc" ]; then
    SHELL_CONFIG="$HOME/.zshrc"
elif [ -n "$BASH_VERSION" ] || [ -f "$HOME/.bashrc" ]; then
    SHELL_CONFIG="$HOME/.bashrc"
else
    SHELL_CONFIG="$HOME/.profile"
fi

EXPORT_LINE="export PATH=\"\$HOME/.larapak/bin:\$PATH\""

if [ -f "$SHELL_CONFIG" ]; then
    if ! grep -q ".larapak/bin" "$SHELL_CONFIG"; then
        echo "" >> "$SHELL_CONFIG"
        echo "# LarapakGo CLI Path" >> "$SHELL_CONFIG"
        echo "$EXPORT_LINE" >> "$SHELL_CONFIG"
        echo "[OK] PATH wis ditambahna maring $SHELL_CONFIG."
    else
        echo "[OK] PATH wis kecathet neng $SHELL_CONFIG."
    fi
fi

# 5. Priksa Instalasi
echo ""
if [ -f "$TARGET_BIN" ]; then
    echo "=========================================================="
    echo " [SUKSES] LarapakGo v$VERSION sukses kepasang neng sistemmu!"
    echo "=========================================================="
    echo ""
    echo "Lokasi eksekusi: $TARGET_BIN"
    echo ""
    echo "Langkah sateruse nggo miwiti:"
    echo "  1. Muat ulang konfigurasi shell:"
    echo "       source $SHELL_CONFIG"
    echo "  2. Tes perintah CLI:"
    echo "       larapak tulung"
    echo "  3. Gawe proyek web anyar nganggo dialek Ngapak:"
    echo "       larapak anyar warung_sate --bahasa=ngapak"
    echo "  4. Mlaku maring folder lan ladeni server:"
    echo "       cd warung_sate"
    echo "       larapak ladeni"
    echo ""
    echo "Ora Ngapak Ora Kepenak!"
    echo ""
else
    echo "[Kliru] Berkas eksekusi ora ditemokna sawise instalasi."
    exit 1
fi
