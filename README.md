# 🐳 GCR — A Minimal Container Runtime in Go

**gcr** is a container runtime built from scratch in Go to demystify how containers work at the syscall level. It now supports rootless networking using `veth` interfaces and internet access via a Linux bridge — all inside a fully reproducible ARM64 environment on Apple Silicon.

---

## 🚀 Quick Start (Apple Silicon)

### 📁 Project Structure
```
project/
├── gcr.lima.yaml # Lima VM config
├── provision.sh # Linux VM provisioning script
└── gcr/ # The container runtime source
```

---

### 🛠️ Prerequisites

Install these on your Mac:

```bash
brew install lima
brew install qemu bash-completion rsync
```

---

### 🔧 Start the Lima VM

```bash
limactl start ./gcr.lima.yaml --name=gcr
```

This will:

- Launch an Ubuntu 24.04 ARM64 VM using Apple's Hypervisor.framework
- Install Go 1.24.3 from go.dev
- Set up bridge interface `br0` + NAT (via iptables)
- Mount your `gcr/` source into the VM at `/home/ubuntu/gcr`

---

### 🐧 SSH into the VM

```bash
limactl shell gcr
```

Inside the VM:

```bash
cd ~/gcr
go version # Should show 1.24.3
go run main.go run ...
```

---

## 🌐 Networking Overview

The provisioned VM includes:

- A **bridge interface `br0`** at `10.0.0.1`
- A **MASQUERADE rule** via iptables for NAT
- IP forwarding enabled
- Containers can connect to the internet via rootless veth setup

You can:

- Use `vishvananda/netlink` to configure veth pairs and routes
- Set DNS manually via mounted `/etc/resolv.conf`

---

## 🔁 Live Code Sync

Any changes made to your `gcr/` directory on the Mac will reflect live inside the VM (`~/gcr`) via Lima's shared mount.

---

## 🧼 Teardown

```bash
limactl stop gcr
limactl delete gcr
```

---

## 🧩 Next Ideas

- Implement `gcr net setup` for auto-veth configuration
- Enable port forwarding (host → container)
- Add support for multi-arch `img2rootfs`

---

## 📬 Feedback & Contributions

Feel free to fork, submit PRs, or open issues. This project is designed for learning, teaching, and hacking.