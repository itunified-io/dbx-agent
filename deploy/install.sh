#!/bin/bash
# dbx-agent installation script
# Usage: sudo ./install.sh [version]
set -euo pipefail

VERSION="${1:-dev}"
INSTALL_DIR="/opt/dbx-agent"
SERVICE_USER="dbx-agent"

echo "=== dbx-agent Installation v${VERSION} ==="

# Create service user
if ! id "${SERVICE_USER}" &>/dev/null; then
    useradd --system --no-create-home --shell /usr/sbin/nologin "${SERVICE_USER}"
    echo "[OK] Created user ${SERVICE_USER}"
fi

# Create directory structure
mkdir -p "${INSTALL_DIR}/bin" "${INSTALL_DIR}/config" "${INSTALL_DIR}/data" \
         "${INSTALL_DIR}/logs" "${INSTALL_DIR}/certs" "${INSTALL_DIR}/templates"

# Copy binary
if [ -f "bin/dbx-agent" ]; then
    cp bin/dbx-agent "${INSTALL_DIR}/bin/dbx-agent"
    chmod 755 "${INSTALL_DIR}/bin/dbx-agent"
    echo "[OK] Installed binary to ${INSTALL_DIR}/bin/dbx-agent"
else
    echo "[SKIP] No binary found in bin/ — build first with 'make build'"
fi

# Copy default config if not exists
if [ ! -f "${INSTALL_DIR}/config/agent.yaml" ]; then
    cat > "${INSTALL_DIR}/config/agent.yaml" <<'YAML'
# dbx-agent configuration
# See https://github.com/itunified-io/dbx-agent for documentation

central:
  url: https://central.example.com:8091
  # ca_fingerprint: "sha256:..."

agent:
  id: "change-me"
  host_port: 9100

tls:
  cert_file: /opt/dbx-agent/certs/agent.crt
  key_file: /opt/dbx-agent/certs/agent.key
  ca_file: /opt/dbx-agent/certs/ca.crt

vault:
  address: https://vault.example.com:8200
  auth_method: approle

sinks:
  victoria_metrics:
    url: https://vm.example.com:8428/api/v1/write
YAML
    echo "[OK] Created default config at ${INSTALL_DIR}/config/agent.yaml"
    echo "     IMPORTANT: Edit agent.yaml before starting the service!"
fi

# Set ownership
chown -R "${SERVICE_USER}:${SERVICE_USER}" "${INSTALL_DIR}"

# Install systemd service
cp deploy/dbx-agent.service /etc/systemd/system/dbx-agent.service
systemctl daemon-reload
echo "[OK] Installed systemd service"

echo ""
echo "=== Installation Complete ==="
echo "1. Edit ${INSTALL_DIR}/config/agent.yaml"
echo "2. Place TLS certificates in ${INSTALL_DIR}/certs/"
echo "3. Start: systemctl start dbx-agent"
echo "4. Enable on boot: systemctl enable dbx-agent"
echo "5. Check status: systemctl status dbx-agent"
echo "6. View logs: journalctl -u dbx-agent -f"
