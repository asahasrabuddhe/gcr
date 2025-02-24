package main

import (
	"crypto/sha256"
	"encoding/hex"
	"log"
	"os"
	"os/exec"
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

	cmd := exec.Command("/proc/self/exe", append([]string{"fork", hash}, os.Args[2:]...)...)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWNS | syscall.CLONE_NEWIPC | syscall.CLONE_NEWPID | syscall.CLONE_NEWUTS | syscall.CLONE_NEWUSER,
		UidMappings: []syscall.SysProcIDMap{
			{ContainerID: 0, HostID: os.Getuid(), Size: 1},
		},
		GidMappings: []syscall.SysProcIDMap{
			{ContainerID: 0, HostID: os.Getgid(), Size: 1},
		},
	}
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		log.Fatal(err)
	}
}

func fork() {
	printIds()

	err := syscall.Sethostname([]byte(os.Args[2]))
	if err != nil {
		log.Fatal(err)
	}

	cmd := exec.Command(os.Args[4], os.Args[5:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err = cmd.Run()
	if err != nil {
		log.Fatal(err)
	}
}

func printIds() {
	log.Printf("running as pid: %d | uid: %d | gid: %d", os.Getpid(), os.Getuid(), os.Getgid())
}
