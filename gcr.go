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

func run() {
	printIds()

	hashBytes := sha256.Sum256([]byte(time.Now().String()))
	hash := hex.EncodeToString(hashBytes[:])
	hash = hash[:12]

	cmd := command("/proc/self/exe", append([]string{"fork", hash}, os.Args[2:]...)...)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWNS | syscall.CLONE_NEWIPC | syscall.CLONE_NEWPID | syscall.CLONE_NEWUTS | syscall.CLONE_NEWUSER,
		UidMappings: []syscall.SysProcIDMap{
			{ContainerID: 0, HostID: os.Getuid(), Size: 1},
		},
		GidMappings: []syscall.SysProcIDMap{
			{ContainerID: 0, HostID: os.Getgid(), Size: 1},
		},
	}
	must(cmd.Run())
}

func fork() {
	printIds()

	must(syscall.Sethostname([]byte(os.Args[2])))

	homeDir, err := os.UserHomeDir()
	must(err)

	newRoot := filepath.Join(homeDir, "rootfs", os.Args[3])
	putOld := filepath.Join(newRoot, ".put_old")
	must(os.MkdirAll(putOld, 0755))
	// change root using Pivot Root
	must(syscall.Mount(newRoot, newRoot, "", syscall.MS_BIND|syscall.MS_REC, ""))
	must(syscall.PivotRoot(newRoot, putOld))
	must(syscall.Chdir("/"))
	// detach the old root
	putOld = filepath.Base(putOld)
	must(syscall.Unmount(putOld, syscall.MNT_DETACH))
	must(os.RemoveAll(putOld))

	must(command(os.Args[4], os.Args[5:]...).Run())
}

func printIds() {
	log.Printf("running as pid: %d | uid: %d | gid: %d", os.Getpid(), os.Getuid(), os.Getgid())
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}

func command(name string, arg ...string) *exec.Cmd {
	cmd := exec.Command(name, arg...)
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
	return cmd
}
