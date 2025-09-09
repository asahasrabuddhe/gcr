#!/bin/bash
set -e

echo "[+] Installing system dependencies..."
apt-get update
apt-get install -y \
  uidmap \
  iproute2 \
  iptables \
  bridge-utils \
  net-tools \
  curl \
  tar \
  git \
  bash-completion \
  g++ \
  gcc \
  make \
  unzip

echo "[+] Enabling unprivileged user namespaces..."
sysctl -w kernel.unprivileged_userns_clone=1
echo "kernel.unprivileged_userns_clone=1" >> /etc/sysctl.conf

echo "[+] Installing Go 1.23.6 from go.dev..."
GO_VERSION=1.23.6
ARCH=arm64
curl -LO https://go.dev/dl/go${GO_VERSION}.linux-${ARCH}.tar.gz
rm -rf /usr/local/go
tar -C /usr/local -xzf go${GO_VERSION}.linux-${ARCH}.tar.gz

cat >> /etc/profile.d/go_env.sh <<EOF
export GOROOT=/usr/local/go
export GOPATH=/home/ubuntu/go
export PATH=\$PATH:\$GOROOT/bin:\$GOPATH/bin
EOF
chmod +x /etc/profile.d/go_env.sh

echo "[+] Setting up Go env and shell completion..."
echo "source /etc/profile.d/go_env.sh" >> /home/ubuntu/.bashrc
echo "source /etc/bash_completion" >> /home/ubuntu/.bashrc
chown -R ubuntu:ubuntu /home/ubuntu

echo "[+] Creating bridge br0..."
if ! ip link show br0 > /dev/null 2>&1; then
  ip link add br0 type bridge
  ip addr add 10.0.0.1/24 dev br0
  ip link set br0 up
fi

echo "[+] Enabling IP forwarding and NAT..."
sysctl -w net.ipv4.ip_forward=1
echo "net.ipv4.ip_forward=1" >> /etc/sysctl.conf
if ! iptables -t nat -C POSTROUTING -s 10.0.0.0/24 ! -o br0 -j MASQUERADE 2>/dev/null; then
  iptables -t nat -A POSTROUTING -s 10.0.0.0/24 ! -o br0 -j MASQUERADE
fi

echo "[✓] Provisioning complete"
