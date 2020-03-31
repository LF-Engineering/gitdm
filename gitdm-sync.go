package main

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"runtime/debug"
	"strings"
	"sync"
	"syscall"
	"time"
)

var gMtx *sync.Mutex

func fatalOnError(err error) {
	if err != nil {
		tm := time.Now()
		fmt.Printf("Error(time=%+v):\nError: '%s'\nStacktrace:\n%s\n", tm, err.Error(), string(debug.Stack()))
		fmt.Fprintf(os.Stderr, "Error(time=%+v):\nError: '%s'\nStacktrace:\n", tm, err.Error())
		panic("stacktrace")
	}
}

func fatalf(f string, a ...interface{}) {
	fatalOnError(fmt.Errorf(f, a...))
}

func execCommand(cmdAndArgs []string, env map[string]string) {
	command := cmdAndArgs[0]
	arguments := cmdAndArgs[1:]
	cmd := exec.Command(command, arguments...)
	if len(env) > 0 {
		newEnv := os.Environ()
		for key, value := range env {
			newEnv = append(newEnv, key+"="+value)
		}
		cmd.Env = newEnv
	}
	fatalOnError(cmd.Start())
	fatalOnError(cmd.Wait())
}

func requestInfo(r *http.Request) string {
	agent := ""
	hdr := r.Header
	if hdr != nil {
		uAgentAry, ok := hdr["User-Agent"]
		if ok {
			agent = strings.Join(uAgentAry, ", ")
		}
	}
	if agent != "" {
		return fmt.Sprintf("IP: %s, agent: %s", r.RemoteAddr, agent)
	}
	return fmt.Sprintf("IP: %s", r.RemoteAddr)
}

func handle(w http.ResponseWriter, req *http.Request) {
	info := requestInfo(req)
	fmt.Printf("Request: %s\n", info)
	var err error
	defer func() {
		fmt.Printf("Request(exit): %s err:%v\n", info, err)
	}()
	execCommand([]string{"rm", "-rf", "gitdm"}, nil)
	defer func() {
		execCommand([]string{"rm", "-rf", "gitdm"}, nil)
	}()
	cmd := []string{
		"git",
		"clone",
		fmt.Sprintf(
			"https://%s:%s@github.com/%s",
			os.Getenv("GITDM_GIT_USER"),
			os.Getenv("GITDM_GIT_OAUTH"),
			os.Getenv("GITDM_GIT_REPO"),
		),
	}
	env := map[string]string{"GIT_TERMINAL_PROMPT": "0"}
	execCommand(cmd, env)
	wd, err := os.Getwd()
	fatalOnError(err)
	gMtx.Lock()
	defer func() {
		gMtx.Unlock()
	}()
	fatalOnError(os.Chdir("gitdm"))
	defer func() {
		_ = os.Chdir(wd)
	}()
}

func checkEnv() {
	requiredEnv := []string{"GITDM_GIT_REPO", "GITDM_GIT_USER", "GITDM_GIT_OAUTH"}
	for _, env := range requiredEnv {
		if os.Getenv(env) == "" {
			fatalf("%s env variable must be set", env)
		}
	}
}

func serve() {
	fmt.Printf("Starting sync server\n")
	checkEnv()
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGUSR1, syscall.SIGALRM)
	go func() {
		for {
			sig := <-sigs
			fmt.Printf("Exiting due to signal %v\n", sig)
			os.Exit(1)
		}
	}()
	gMtx = &sync.Mutex{}
	http.HandleFunc("/", handle)
	fatalOnError(http.ListenAndServe("0.0.0.0:7070", nil))
}

func main() {
	serve()
	fatalf("serve exited without error, returning error state anyway")
}
