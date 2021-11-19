package main

import (
	"bytes"
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"

	"github.com/mattn/go-isatty"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

const name = "kubectl-shell"

const version = "0.0.1"

var revision = "HEAD"

func listPods() error {
	var kubeconfig *string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()

	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		return err
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return err
	}

	pods, err := clientset.CoreV1().Pods("default").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return err
	}

	for _, pod := range pods.Items {
		fmt.Println(pod.GetName())
	}
	return nil
}

func choice() (string, error) {
	exe, err := os.Executable()
	if err != nil {
		return "", err
	}

	cmd := exec.Command("fzf", "--ansi", "--no-preview")
	var out bytes.Buffer
	cmd.Stdin = os.Stdin
	cmd.Stdout = &out

	cmd.Env = append(os.Environ(),
		fmt.Sprintf("FZF_DEFAULT_COMMAND=%s", exe),
		fmt.Sprintf("_KUBECTX_FORCE_COLOR=1"))
	if err := cmd.Run(); err != nil {
		if _, ok := err.(*exec.ExitError); !ok {
			return "", err
		}
	}
	choice := strings.TrimSpace(out.String())
	if choice == "" {
		return "", errors.New("you did not choose any of the options")
	}
	return choice, nil
}

func fzfInstalled() bool {
	v, _ := exec.LookPath("fzf")
	return v != ""
}

func main() {
	var shell string
	var showVersion bool
	flag.StringVar(&shell, "e", "/bin/bash", "Used shell")
	flag.BoolVar(&showVersion, "V", false, "Print the version")
	flag.Parse()

	if showVersion {
		fmt.Printf("%s %s (rev: %s/%s)\n", name, version, revision, runtime.Version())
		return
	}

	pod := flag.Arg(0)
	if pod == "" && os.Getenv("FZF_DEFAULT_COMMAND") == "" && isatty.IsTerminal(os.Stdout.Fd()) && fzfInstalled() {
		c, err := choice()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		pod = c
	}

	if pod == "" {
		if err := listPods(); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		return
	}

	args := []string{
		"exec",
		"--stdin",
		"--tty",
		pod,
		"--",
	}
	if flag.NArg() > 0 {
		args = append(args, flag.Args()[1:]...)
	}
	if len(args) == 5 {
		args = append(args, shell)
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
