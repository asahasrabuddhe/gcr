package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
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
	printIds("run")
	// Why do we fork here?
	// ---------------------
	// Namespaces in Linux are tied to processes — you can only "live inside"
	// a namespace if you start execution there. Simply calling unshare()
	// in the current process won’t give full isolation (e.g. a PID namespace
	// still shows the old PID numbers).
	//
	// By re-executing /proc/self/exe with the "fork" subcommand, we create
	// a child process inside the new namespace (via clone flags). This child
	// becomes PID 1 in that namespace — the "init" process responsible for
	// managing it. Without this fork/exec trick, the container illusion
	// wouldn’t work correctly.
	cmd := command("/proc/self/exe", append([]string{"fork"}, os.Args[2:]...)...)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWPID,
	}
	must(cmd.Run())
}

func fork() {
	printIds("fork")
	fmt.Println("running command:", os.Args[2:])

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	fmt.Println("Container running... press Ctrl+C to stop.")
	<-sigs
	fmt.Println("Container shutting down.")
}

func printIds(fn string) {
	log.Printf("[%s] as pid %d", fn, os.Getpid())
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
