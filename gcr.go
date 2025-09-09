package main

import (
	"crypto/sha256"
	"encoding/hex"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"time"
)

// gcr run <image> <command>
func main() {
	if len(os.Args) < 3 {
		log.Fatal("Usage: gcr run <image> <command>")
	}

	switch os.Args[1] {
	case "run":
		run()
	case "fork":
		fork()
	default:
		log.Fatal("Usage: gcr run <image> <command>")
	}
}

// run re-executes the current binary with the "fork" subcommand, creating a
// child process inside a new namespace (PID and mount). This child becomes
// PID 1 in that namespace — the "init" process responsible for managing it.
//
// The reason for this fork/exec trick is that namespaces in Linux are tied
// to processes: you can only "live inside" a namespace if you start execution
// there. Simply calling unshare() in the current process won't give full
// isolation (e.g. a PID namespace still shows the old PID numbers).
func run() {
	printIds("run")

	// generate container id
	hashBytes := sha256.Sum256([]byte(time.Now().String()))
	hash := hex.EncodeToString(hashBytes[:])
	hash = hash[:12]

	cmd := command("/proc/self/exe", append([]string{"fork", hash}, os.Args[2:]...)...)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWPID | syscall.CLONE_NEWNS | syscall.CLONE_NEWUSER | syscall.CLONE_NEWUTS | syscall.CLONE_NEWIPC,
		UidMappings: []syscall.SysProcIDMap{
			{ContainerID: 0, HostID: os.Getuid(), Size: 1},
		},
		GidMappings: []syscall.SysProcIDMap{
			{ContainerID: 0, HostID: os.Getgid(), Size: 1},
		},
		GidMappingsEnableSetgroups: false,
		Credential: &syscall.Credential{
			Uid: uint32(0),
			Gid: uint32(0),
		},
	}
	must(cmd.Run())
}

// fork is the entrypoint for the child process created with CLONE_NEWPID |
// CLONE_NEWNS. It will act as PID 1 in the container after pivot_root
// succeeds.
//
// We are now inside the child process created with CLONE_NEWNS (new mount namespace)
// and (likely) CLONE_NEWPID (new PID namespace). This process will act as PID 1
// inside the container after pivot_root succeeds.
func fork() {
	printIds("fork")

	// set hostname
	must(syscall.Sethostname([]byte(os.Args[2])))

	// Get current working directory (host side), which we’ll use to resolve the rootfs path.
	dir, err := os.Getwd()
	must(err)

	// --- Step 1: Compute paths ---
	// newRoot: the path to the container’s root filesystem (extracted image).
	// putOld: a temporary directory inside newRoot where the old root will be parked during pivot_root.
	newRoot := filepath.Join(dir, "rootfs", os.Args[3])
	putOld := filepath.Join(newRoot, ".put_old")
	must(os.MkdirAll(putOld, 0700)) // must exist before pivot_root can succeed

	// --- Step 2: Prepare essential mounts ---
	// Containers need minimal system mounts like /proc, /dev, and /tmp.
	// We set them up *inside newRoot* so that once we pivot, they’re visible to the container.
	//
	// Use `defer` to clean up before exiting.
	defer mount("proc", filepath.Join(newRoot, "proc"), "proc",
		syscall.MS_NOEXEC|syscall.MS_NOSUID|syscall.MS_NODEV, "")()
	defer mount("tmpfs", filepath.Join(newRoot, "dev"), "tmpfs",
		syscall.MS_NOSUID|syscall.MS_STRICTATIME, "mode=755,size=65536k")()
	defer mount("tmpfs", filepath.Join(newRoot, "tmp"), "tmpfs",
		syscall.MS_NOSUID|syscall.MS_STRICTATIME, "mode=755,size=65536k")()

	// --- Step 3: Isolate mount namespace ---
	// Make mount propagation private: prevents mount/unmount changes inside this namespace
	// from leaking back to the host (or vice versa).
	must(syscall.Mount("", "/", "", syscall.MS_REC|syscall.MS_PRIVATE, ""))
	// Ensure newRoot is a distinct mount point (required for pivot_root).
	must(syscall.Mount(newRoot, newRoot, "", syscall.MS_BIND|syscall.MS_REC, ""))

	// --- Step 4: Pivot into the new root ---
	// Switch our CWD into newRoot. pivot_root requires both newRoot and putOld
	// to be inside the current rootfs.
	must(syscall.Chdir(newRoot))

	// Replace the current rootfs with newRoot, moving the old root to putOld.
	must(syscall.PivotRoot(newRoot, putOld))

	// --- Step 5: Finalise new root ---
	// Move into the new root (‘/’) so relative paths resolve correctly now.
	must(os.Chdir("/"))

	// Detach and remove the old root. After this, the container cannot “escape”
	// back to the host filesystem.
	putOld = filepath.Base(putOld) // now visible inside the new namespace as /.put_old
	must(syscall.Unmount(putOld, syscall.MNT_DETACH))
	must(os.RemoveAll(putOld))

	// --- Step 6: Run the container’s init command ---
	// At this point, the process is PID 1 in the new PID namespace,
	// with its own root filesystem. Executing the target command
	// effectively “starts” the container’s workload.
	must(command(os.Args[4], os.Args[5:]...).Run())
}

// printIds logs the current process ID, annotated with the given function name.
//
// printIds is used to log process IDs at key points in the container creation
// process.
func printIds(fn string) {
	log.Printf("[%s] as pid %d", fn, os.Getpid())
}

// must is a convenience function for aborting the program if an error is
// present. It should be used to wrap functions that return an error, and
// provides a nicer error message than the standard library's log.Fatal.
func must(err error) {
	if err != nil {
		panic(err)
	}
}

// command wraps os/exec.Command to create an exec.Cmd with the given name and arguments,
// and sets the Stdin, Stdout, and Stderr fields of the Cmd to the corresponding
// os.Stdin, os.Stdout, and os.Stderr values.
//
// The returned Cmd is suitable for direct use with Run, Start, etc.
func command(name string, arg ...string) *exec.Cmd {
	cmd := exec.Command(name, arg...)
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
	return cmd
}

// mount wraps syscall.Mount to create a mount point at the given path,
// using the given source, filesystem type, mount flags, and data.
//
// The returned function is a closure that reverses the mount operation,
// suitable for use with defer.
func mount(source, path, fstype string, flags uintptr, data string) func() {
	must(syscall.Mount(source, path, fstype, flags, data))
	return func() { must(syscall.Unmount("/"+filepath.Base(path), 0)) }
}
