package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"runtime/debug"
	"strings"
	"sync"
	"syscall"
	"time"

	yaml "gopkg.in/yaml.v2"
)

var gMtx *sync.Mutex

type enrollmentShortOutput struct {
	End          string `yaml:"T"`
	Organization string `yaml:"C"`
	Start        string `yaml:"F"`
}

type identityShortOutput struct {
	Email    *string `yaml:"E,omitempty"`
	Name     *string `yaml:"M,omitempty"`
	Source   string  `yaml:"S"`
	Username *string `yaml:"U,omitempty"`
}

type allOutput struct {
	CountryCode *string                  `yaml:"C,omitempty"`
	Email       *string                  `yaml:"E,omitempty"`
	Enrollments []*enrollmentShortOutput `yaml:"R,omitempty"`
	Gender      *string                  `yaml:"S,omitempty"`
	Identities  []*identityShortOutput   `yaml:"I,omitempty"`
	IsBot       *int64                   `yaml:"B,omitempty"`
	Name        *string                  `yaml:"U,omitempty"`
}

type allArrayOutput struct {
	Profiles []*allOutput `yaml:"P,omitempty"`
}

func (e *enrollmentShortOutput) size() int {
	return 48 + len([]byte(e.Organization))
}

func (i *identityShortOutput) size() int {
	b := 8 + len(i.Source)
	if i.Name != nil {
		b += 8 + len([]byte(*i.Name))
	}
	if i.Email != nil {
		b += 8 + len([]byte(*i.Email))
	}
	if i.Username != nil {
		b += 8 + len([]byte(*i.Username))
	}
	return b
}

func (a *allOutput) size() int {
	b := 12
	if a.CountryCode != nil {
		b += 6 + len(*a.CountryCode)
	}
	if a.Gender != nil {
		b += 6 + len(*a.Gender)
	}
	if a.Email != nil {
		b += 6 + len([]byte(*a.Email))
	}
	if a.Name != nil {
		b += 6 + len([]byte(*a.Name))
	}
	if a.IsBot != nil {
		b += 7
	}
	for _, identity := range a.Identities {
		b += identity.size()
	}
	for _, enrollment := range a.Enrollments {
		b += enrollment.size()
	}
	return b
}

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

func processRepo() {
	i := 1
	var profs []*allOutput
	for {
		data, err := ioutil.ReadFile(fmt.Sprintf("profiles%d.yaml", i))
		if err != nil {
			break
		}
		var all allArrayOutput
		fatalOnError(yaml.Unmarshal(data, &all))
		profs = append(profs, all.Profiles...)
		i++
	}
	currSize := 0
	profSize := 0
	from := 0
	maxSize := (1 << 20) - 8
	ranges := [][2]int{}
	for i, prof := range profs {
		profSize = prof.size()
		if currSize+profSize > maxSize {
			ranges = append(ranges, [2]int{from, i})
			from = i
			currSize = 0
		} else {
			currSize += profSize
		}
	}
  if from != len(profs)-1 {
		ranges = append(ranges, [2]int{from, len(profs)})
  }
  for i, rng := range ranges {
		var all allArrayOutput
    all.Profiles = profs[rng[0]:rng[1]]
    data, err := yaml.Marshal(&all)
    fatalOnError(err)
    fatalOnError(ioutil.WriteFile(fmt.Sprintf("profiles%d.yaml", i+1), data, 0644))
  }
  fmt.Printf("Written %d profile files\n", len(ranges))
	execCommand([]string{"git", "add", "."}, nil)
	execCommand(
    []string{
      "git",
      "commit",
      "-asm",
      "--author",
      os.Getenv("GITDM_GIT_USER"),
      fmt.Sprintf("gitdm-sync @ %s", time.Now.Format("2006-01-02 15:04:05")),
    },
    nil,
  )
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
	processRepo()
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
