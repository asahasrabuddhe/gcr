# gcr

A minimal container runtime written in **Go**, designed to help developers understand **how Docker works internally**. This project serves as a **learning tool** and a **companion** to my presentation on container internals, focusing on **Linux namespaces**, **pivot_root**, and process isolation.

## Features

* Implements **PID, UTS, IPC, USER, and MNT namespaces**
* Uses `pivot_root` to isolate the filesystem
* Mounts essential filesystems (`proc`, `tmpfs`, etc.)
* Generates unique container IDs using **SHA-256 hashing**
* Supports basic command execution inside containers

## Installation

Ensure you have **Go 1.16+** installed:

```sh
# Clone the repository
git clone https://github.com/asahasrabuddhe/gcr.git
cd gcr-lite

# Build the binary
go build -o gcr
```

## Usage

### Run a command inside a container
```sh
./gcr run ubuntu bash
```

- This mounts the Ubuntu file system as root and executes `/bin/sh` inside an isolated environment.
- The container runs with **separate namespaces** and **pivot_root**, ensuring filesystem and process isolation.

### Example Output
```
2025/03/16 08:56:43 running as pid: 4776 | uid: 1000 | gid: 1000
2025/03/16 08:56:43 running as pid: 1 | uid: 0 | gid: 0
root@dda93bff2ea0:/# 
```

### Extract the RootFS of a Docker Image
The img2rootfs tool allows extracting the root filesystem of any Docker image.
```bash
sudo ./img2rootfs -image ubuntu:20.04 -output ~/rootfs/ubuntu2004/
```

* This command pulls the ubuntu:20.04 Docker image and extracts its root filesystem to ~/rootfs/ubuntu2004/.

* The extracted rootfs can be used for further analysis or container experiments.

## How It Works
1. The `run` command forks a child process with **new namespaces**.
2. It sets up **root filesystem isolation** using `pivot_root`.
3. Essential filesystems (`proc`, `dev`, `tmp`) are mounted.
4. The specified command is executed **inside the container**.

## Future Enhancements
- Integrate **cgroups** for resource management
- Add **network namespaces** for network isolation

## License

This project is licensed under the **MIT License**. See [LICENSE](LICENSE) for details.

---

💡 **Why This Project?**

This project is built to **demystify Docker** and help developers understand container internals by experimenting with Go. If you find it useful, ⭐️ the repo and share it with others!

🚀 Happy Hacking!

