package main

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"runtime/debug"
	"strings"
	"syscall"
	"time"
)

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
	http.HandleFunc("/", handle)
	fatalOnError(http.ListenAndServe("0.0.0.0:7070", nil))
}

func main() {
	serve()
	fatalf("serve exited without error, returning error state anyway")
}
