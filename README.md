# 🐳 GCR — A Minimal Container Runtime in Go

**GCR** (Go Container Runtime) is a minimal container runtime built from scratch in Go. It demonstrates how Linux namespaces, cgroups, and chroot work together to provide container isolation. This project is designed for educational purposes to help understand the fundamentals of container runtimes like Docker and containerd.

## 📁 Project Structure

```
gcr/
├── gcr.go              # Main container runtime implementation
├── gcr.lima.yaml       # Lima VM configuration for development
├── provision.sh        # VM provisioning script
├── img2rootfs/         # Tool to convert Docker images to rootfs
│   └── img2rootfs.go   # Docker image to rootfs converter
├── rootfs/             # Example root filesystems for containers
└── go.mod              # Go module definition
```

## 🚀 Features

- Process isolation using Linux namespaces (PID, UTS, IPC, mount, user, network)
- Filesystem isolation using `pivot_root`
- Basic container networking setup
- Support for running containers from root filesystems

## 🛠️ Prerequisites

### For Development (macOS with Apple Silicon)

```bash
# Install Lima for Linux VMs
brew install lima

# Install QEMU and other dependencies
brew install qemu bash-completion rsync
```

### For Running on Linux

- Go 1.25 or later
- Linux kernel with namespaces and cgroups support
- Root privileges (for most operations)

## 🚀 Quick Start

### 1. Start the Development VM

```bash
limactl start gcr.lima.yaml --name=gcr
```

This will:
- Create an Ubuntu 22.04 VM
- Install Go 1.25.0 and dependencies
- Set up a shared directory at `~/gcr`

### 2. SSH into the VM

```bash
limactl shell gcr
```

### 3. Build and Run

Inside the VM:

```bash
cd ~/gcr

# Build the runtime
go build -o gcr .

# Run a container with a shell
sudo ./gcr run <rootfs-name> <command>
```

## Building a Root Filesystem

```bash
# Build the img2rootfs tool
cd img2rootfs
go build -o img2rootfs .

# Convert a Docker image to rootfs
sudo ./img2rootfs -image ubuntu:latest -output ../rootfs/ubuntu
```

## 🔧 How It Works

GCR uses several Linux features to provide container isolation:

1. **Namespaces**: Isolate processes, network, filesystem, etc.
2. **chroot/pivot_root**: Isolate filesystem access

## 🧪 Testing Commands

### 1. PID Namespace

```bash
git checkout v0.1

# Build the runtime
go build -o gcr .

# Run a container with a shell
sudo ./gcr run ubuntu bash
```

#### Output

```
2025/09/09 19:37:04 [run] as pid 5293
2025/09/09 19:37:04 [fork] as pid 1
running command: [ubuntu bash]
Container running... press Ctrl+C to stop.
```

We can see that the initial process (PID 5293) is running in the main namespace (PID 1). The container process (PID 1) is the "init" process, responsible for managing the container.

### 2. MNT Namespace

```bash
git checkout v0.2

# Build the runtime
go build -o gcr .

# Run a container with a shell
sudo ./gcr run ubuntu bash
```

#### Output

```
2025/09/09 19:39:02 [run] as pid 5341
2025/09/09 19:39:02 [fork] as pid 1
root@lima-gcr:/# ls
bin  boot  dev  etc  home  lib  media  mnt  opt  proc  root  run  sbin  srv  sys  tmp  usr  var
root@lima-gcr:/# 
```

We now have filesystem isolation! The container process (PID 1) is running in the root namespace (PID 0). The container process is running in the `/` directory, which is the root of the host filesystem.

### 3. USER Namespace

```bash
git checkout v0.3

# Build the runtime
go build -o gcr .

# Run a container with a shell
sudo ./gcr run ubuntu bash 
```

#### Output

```
2025/09/09 19:40:51 [run] as pid 5391
2025/09/09 19:40:51 [fork] as pid 1
root@lima-gcr:/# ls
bin  boot  dev  etc  home  lib  media  mnt  opt  proc  root  run  sbin  srv  sys  tmp  usr  var
root@lima-gcr:/# 
```

We now have user isolation! The container process (PID 1) is running as the `root` user. We also no longer have to use `sudo` to start our container.

### 4. UTS Namespace

```
ajitem@lima-gcr:~/gcr$ hostname
lima-gcr
ajitem@lima-gcr:~/gcr$ ./gcr run ubuntu bash
2025/09/09 19:44:28 [run] as pid 5412
2025/09/09 19:44:28 [fork] as pid 1
root@lima-gcr:/# hostname
lima-gcr
root@lima-gcr:/# 
```

Notice how the container process (PID 1) is running on the same hostname as the host. The UTS namespace is used to isolate the hostname of the container process.

```bash
git checkout v0.4

# Build the runtime
go build -o gcr .
```

#### Output

```
ajitem@lima-gcr:~/gcr$ hostname
lima-gcr
ajitem@lima-gcr:~/gcr$ ./gcr run ubuntu bash
2025/09/09 19:46:02 [run] as pid 5458
2025/09/09 19:46:02 [fork] as pid 1
root@72741c42c0d4:/# hostname
72741c42c0d4
root@72741c42c0d4:/# 
```

We now have hostname isolation! The container process (PID 1) is running on a different hostname than the host.

### 5. IPC Namespace

```
ajitem@lima-gcr:~/gcr$ ipcmk -M 1M
Shared memory id: 0
ajitem@lima-gcr:~/gcr$ ipcs -m

------ Shared Memory Segments --------
key        shmid      owner      perms      bytes      nattch     status      
0x6393bcb2 0          ajitem     644        1048576    0                       

ajitem@lima-gcr:~/gcr$ ./gcr run ubuntu bash
2025/09/09 19:49:31 [run] as pid 5484
2025/09/09 19:49:31 [fork] as pid 1
root@cc73b5349ac1:/# ipcs -m

------ Shared Memory Segments --------
key        shmid      owner      perms      bytes      nattch     status      
0x6393bcb2 0          root       644        1048576    0                       

root@cc73b5349ac1:/# 
```

To test IPC isolation, we first create a shared memory segment on the host. Then we run a container and right now, the shared memory segment is visible inside the container. The IPC namespace will allow us to isolage this shared memory segment.

```bash
git checkout v0.5

# Build the runtime
go build -o gcr .
```

#### Output

```
ajitem@lima-gcr:~/gcr$ ipcs -m

------ Shared Memory Segments --------
key        shmid      owner      perms      bytes      nattch     status      
0x6393bcb2 2          ajitem     644        1048576    0                       

ajitem@lima-gcr:~/gcr$ ./gcr run ubuntu bash
2025/09/09 19:51:32 [run] as pid 5532
2025/09/09 19:51:32 [fork] as pid 1
root@ee6607005f0b:/# ipcs -m

------ Shared Memory Segments --------
key        shmid      owner      perms      bytes      nattch     status      

root@ee6607005f0b:/# exit
root@ee6607005f0b:/# exit
exit
ajitem@lima-gcr:~/gcr$ ipcrm -m 2
```

Now the shared memory segment is no longer visible inside the container. We need to clear the shared memory segment after we are done with it.

### 6. cgroups

```
ajitem@lima-gcr:~/gcr$ ./cgroups.sh 
>>> Creating cgroup at /sys/fs/cgroup/user.slice/user-501.slice/user@501.service/app.slice/myapp.scope
+pids
5
2019
>>> Spawning background processes
Spawned Process 1 with PID
Spawned Process 2 with PID
Spawned Process 3 with PID
./cgroups.sh: fork: retry: Resource temporarily unavailable
./cgroups.sh: fork: retry: Resource temporarily unavailable
./cgroups.sh: fork: retry: Resource temporarily unavailable
Spawned Process 4 with PID
Spawned Process 5 with PID
Spawned Process 6 with PID
./cgroups.sh: fork: retry: Resource temporarily unavailable
./cgroups.sh: fork: retry: Resource temporarily unavailable
./cgroups.sh: fork: retry: Resource temporarily unavailable
Spawned Process 7 with PID
Spawned Process 8 with PID
Spawned Process 9 with PID
./cgroups.sh: fork: retry: Resource temporarily unavailable
./cgroups.sh: fork: retry: Resource temporarily unavailable
./cgroups.sh: fork: retry: Resource temporarily unavailable
Spawned Process 10 with PID
>>> Status
pids.current = 3
cgroup.procs:
2019
2073
2075
>>> Sleeping for 5 seconds before cleanup...
>>> Restoring shell back to original cgroup: /user.slice/user-501.slice/session-2.scope
2019
>>> Cleaning up /sys/fs/cgroup/user.slice/user-501.slice/user@501.service/app.slice/myapp.scope
```

The cgroups will allow us to limit the number of processes that can run inside the container. The keen-eyed might notice that despite the `pids.max` is set to 5, only 3 processes run at a time. This is because the bash process (PID 5) and the process that runs the for loop (PID 2019) take up two slots leaving space for only 3 processes to run. 

## 📬 Feedback & Contributions

Contributions are welcome! Please feel free to submit issues or pull requests.

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
