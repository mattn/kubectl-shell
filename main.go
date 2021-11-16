package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"syscall"
)

const name = "kubectl-shell"

const version = "0.0.1"

var revision = "HEAD"

func main() {
	var showVersion bool
	flag.BoolVar(&showVersion, "V", false, "Print the version")
	flag.Parse()

	if showVersion {
		fmt.Printf("%s %s (rev: %s/%s)\n", name, version, revision, runtime.Version())
		return
	}

	if flag.NArg() == 0 {
		fmt.Fprintln(os.Stderr, "No pod given")
		os.Exit(1)
	}
	pod := flag.Arg(0)
	args := []string{
		"exec",
		"--stdin",
		"--tty",
		pod,
		"--",
	}
	args = append(args, flag.Args()[1:]...)
	if len(args) == 5 {
		args = append(args, "/bin/bash")
	}
	cmd := exec.Command("kubectl", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	if err := cmd.Run(); err != nil {
		if e2, ok := err.(*exec.ExitError); ok {
			if s, ok := e2.Sys().(syscall.WaitStatus); ok {
				os.Exit(s.ExitStatus())
			} else {
				panic(errors.New("Unimplemented for system where exec.ExitError.Sys() is not syscall.WaitStatus."))
			}
		}
	}
}
