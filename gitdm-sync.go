package main

import (
	"bytes"
	"fmt"
	"html"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"runtime/debug"
	"sort"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/pkg/errors"
	yaml "gopkg.in/yaml.v2"
)

var (
	gMtx *sync.Mutex
	gw   http.ResponseWriter
)

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

func fatalOnError(err error, pnic bool) bool {
	if err != nil {
		tm := time.Now()
		fmt.Printf("Error(time=%+v):\nError: '%s'\nStacktrace:\n%s\n", tm, err.Error(), string(debug.Stack()))
		fmt.Fprintf(os.Stderr, "Error(time=%+v):\nError: '%s'\nStacktrace:\n", tm, err.Error())
		gw.WriteHeader(http.StatusBadRequest)
		_, _ = io.WriteString(gw, err.Error()+"\n")
		if pnic {
			panic("stacktrace")
		}
		return true
	}
	return false
}

func fatalf(pnic bool, f string, a ...interface{}) {
	fatalOnError(fmt.Errorf(f, a...), pnic)
}

func execCommand(cmdAndArgs []string, env map[string]string, dbg int, allowedExitCodes []int) (string, bool) {
	if dbg > 0 {
		if len(env) > 0 {
			fmt.Printf("%+v %s\n", env, strings.Join(cmdAndArgs, " "))
		} else {
			fmt.Printf("%s\n", strings.Join(cmdAndArgs, " "))
		}
	}
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
	var (
		stdOut bytes.Buffer
		stdErr bytes.Buffer
	)
	cmd.Stderr = &stdErr
	cmd.Stdout = &stdOut
	if fatalOnError(cmd.Start(), false) {
		return "cmd.Start() failed", false
	}
	err := cmd.Wait()
	if err != nil {
		for _, allowed := range allowedExitCodes {
			if err.Error() == fmt.Sprintf("exit status %d", allowed) {
				if dbg > 0 {
					fmt.Printf("exit code %d but this is allowed\n", allowed)
				}
				err = nil
				break
			}
		}
	}
	if err != nil || dbg > 1 {
		outStr := stdOut.String()
		errStr := stdErr.String()
		fmt.Printf("STDOUT:\n%v\n", outStr)
		fmt.Printf("STDERR:\n%v\n", errStr)
		if err != nil {
			err = fmt.Errorf("%+v\nstdout:\n%s\nstderr:\n%s", err, outStr, errStr)
		}
		if fatalOnError(err, false) {
			return "cmd.Wait() failed", false
		}
	}
	return stdOut.String(), true
}

func requestInfo(r *http.Request) string {
	agent := ""
	hdr := r.Header
	method := r.Method
	path := html.EscapeString(r.URL.Path)
	if hdr != nil {
		uAgentAry, ok := hdr["User-Agent"]
		if ok {
			agent = strings.Join(uAgentAry, ", ")
		}
	}
	if agent != "" {
		return fmt.Sprintf("IP: %s, agent: %s, method: %s, path: %s", r.RemoteAddr, agent, method, path)
	}
	return fmt.Sprintf("IP: %s, method: %s, path: %s", r.RemoteAddr, method, path)
}

func syncRepo() bool {
	i := 1
	var profs []*allOutput
	for {
		fmt.Printf("reading profiles%d.yaml\n", i)
		data, err := ioutil.ReadFile(fmt.Sprintf("profiles%d.yaml", i))
		if err != nil {
			break
		}
		var all allArrayOutput
		fmt.Printf("parse profiles%d.yaml\n", i)
		err = yaml.Unmarshal(data, &all)
		if err != nil {
			err = errors.Wrap(err, fmt.Sprintf("profiles%d.yaml", i))
		}
		if fatalOnError(err, false) {
			return false
		}
		profs = append(profs, all.Profiles...)
		i++
	}
	fmt.Printf("sorting\n")
	sort.SliceStable(profs, func(i, j int) bool {
		iS := ""
		if profs[i].Name != nil {
			iS += ":" + *(profs[i].Name)
		}
		if profs[i].Email != nil {
			iS += ":" + *(profs[i].Email)
		}
		if profs[i].CountryCode != nil {
			iS += ":" + *(profs[i].CountryCode)
		}
		if profs[i].Gender != nil {
			iS += ":" + *(profs[i].Gender)
		}
		if profs[i].IsBot != nil {
			if *(profs[i].IsBot) == 0 {
				iS += ":0"
			} else {
				iS += ":1"
			}
		}
		jS := ""
		if profs[j].Name != nil {
			jS += ":" + *(profs[j].Name)
		}
		if profs[j].Email != nil {
			jS += ":" + *(profs[j].Email)
		}
		if profs[j].CountryCode != nil {
			jS += ":" + *(profs[j].CountryCode)
		}
		if profs[j].Gender != nil {
			jS += ":" + *(profs[j].Gender)
		}
		if profs[j].IsBot != nil {
			if *(profs[j].IsBot) == 0 {
				jS += ":0"
			} else {
				jS += ":1"
			}
		}
		return iS < jS
	})
	for k := range profs {
		if len(profs[k].Enrollments) > 1 {
			sort.SliceStable(profs[k].Enrollments, func(i, j int) bool {
				rols := profs[k].Enrollments
				if rols[i].Start == rols[j].Start {
					if rols[i].End == rols[j].End {
						return rols[i].Organization < rols[j].Organization
					}
					return rols[i].End < rols[j].End
				}
				return rols[i].Start < rols[j].Start
			})
		}
		if len(profs[k].Identities) > 1 {
			sort.SliceStable(profs[k].Identities, func(i, j int) bool {
				ids := profs[k].Identities
				iS := ids[i].Source
				if ids[i].Name != nil {
					iS += ":" + *(ids[i].Name)
				}
				if ids[i].Email != nil {
					iS += ":" + *(ids[i].Email)
				}
				if ids[i].Username != nil {
					iS += ":" + *(ids[i].Username)
				}
				jS := ids[j].Source
				if ids[j].Name != nil {
					jS += ":" + *(ids[j].Name)
				}
				if ids[j].Email != nil {
					jS += ":" + *(ids[j].Email)
				}
				if ids[j].Username != nil {
					jS += ":" + *(ids[j].Username)
				}
				return iS < jS
			})
		}
	}
	currSize := 0
	profSize := 0
	from := 0
	maxSize := (1 << 20) - 8
	ranges := [][2]int{}
	fmt.Printf("fitting %d profs in files no larger than %d bytes\n", len(profs), maxSize)
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
	fmt.Printf("data ranges: %+v\n", ranges)
	for i, rng := range ranges {
		var all allArrayOutput
		all.Profiles = profs[rng[0]:rng[1]]
		fmt.Printf("writting profiles%d.yaml [%d-%d]\n", i+1, rng[0], rng[1])
		data, err := yaml.Marshal(&all)
		if fatalOnError(err, false) {
			return false
		}
		if fatalOnError(ioutil.WriteFile(fmt.Sprintf("profiles%d.yaml", i+1), data, 0644), false) {
			return false
		}
	}
	fmt.Printf("written %d profile files\n", len(ranges))
	fmt.Printf("git status *.yaml\n")
	status, ok := execCommand([]string{"git", "status", "*.yaml"}, nil, 1, []int{})
	if !ok {
		return false
	}
	if strings.Contains(status, "nothing to commit, working tree clean") {
		fmt.Printf("Profile YAML files don't need updates\n")
		return true
	}
	fmt.Printf("git add *.yaml\n")
	_, ok = execCommand([]string{"git", "add", "*.yaml"}, nil, 1, []int{})
	if !ok {
		return false
	}
	fmt.Printf("git config user.name get\n")
	cfg, ok := execCommand([]string{"git", "config", "--global", "user.name"}, nil, 1, []int{1})
	if !ok {
		return false
	}
	if strings.TrimSpace(cfg) == "" {
		fmt.Printf("git config user.name set\n")
		_, ok = execCommand([]string{"git", "config", "--global", "user.name", os.Getenv("GITDM_GIT_USER")}, nil, 0, []int{})
		if !ok {
			return false
		}
	}
	fmt.Printf("git config user.email get\n")
	cfg, ok = execCommand([]string{"git", "config", "--global", "user.email"}, nil, 1, []int{1})
	if !ok {
		return false
	}
	if strings.TrimSpace(cfg) == "" {
		fmt.Printf("git config user.email set\n")
		_, ok = execCommand([]string{"git", "config", "--global", "user.email", os.Getenv("GITDM_GIT_EMAIL")}, nil, 0, []int{})
		if !ok {
			return false
		}
	}
	fmt.Printf("git commit\n")
	_, ok = execCommand(
		[]string{
			"git",
			"commit",
			"-sm",
			fmt.Sprintf("%s gitdm-sync @ %s", os.Getenv("GITDM_GITHUB_USER"), time.Now().Format("2006-01-02 15:04:05")),
		},
		nil,
		1,
		[]int{},
	)
	if !ok {
		return false
	}
	fmt.Printf("git push\n")
	_, ok = execCommand(
		[]string{
			"git",
			"push",
			"--repo",
			fmt.Sprintf(
				"https://%s:%s@github.com/%s",
				os.Getenv("GITDM_GITHUB_USER"),
				os.Getenv("GITDM_GITHUB_OAUTH"),
				os.Getenv("GITDM_GITHUB_REPO"),
			),
		},
		nil,
		0,
		[]int{},
	)
	if !ok {
		return false
	}
	fmt.Printf("processing repo finished\n")
	return true
}

func checkRepo() bool {
	i := 1
	for {
		fmt.Printf("reading profiles%d.yaml\n", i)
		data, err := ioutil.ReadFile(fmt.Sprintf("profiles%d.yaml", i))
		if err != nil {
			break
		}
		var all allArrayOutput
		fmt.Printf("parse profiles%d.yaml\n", i)
		err = yaml.Unmarshal(data, &all)
		if err != nil {
			err = errors.Wrap(err, fmt.Sprintf("profiles%d.yaml", i))
		}
		if fatalOnError(err, false) {
			return false
		}
		i++
	}
	fmt.Printf("checking repo finished\n")
	return true
}

func handlePush(w http.ResponseWriter, req *http.Request) {
	info := requestInfo(req)
	fmt.Printf("Request: %s\n", info)
	var err error
	defer func() {
		fmt.Printf("Request(exit): %s err:%v\n", info, err)
	}()
	fmt.Printf("lock mutex\n")
	gMtx.Lock()
	defer func() {
		fmt.Printf("unlock mutex\n")
		gMtx.Unlock()
	}()
	gw = w
	fmt.Printf("Cleanup repo before\n")
	execCommand([]string{"rm", "-rf", "gitdm"}, nil, 1, []int{})
	defer func() {
		fmt.Printf("Cleanup repo after\n")
		execCommand([]string{"rm", "-rf", "gitdm"}, nil, 1, []int{})
	}()
	fmt.Printf("git clone\n")
	cmd := []string{
		"git",
		"clone",
		fmt.Sprintf(
			"https://%s:%s@github.com/%s",
			os.Getenv("GITDM_GITHUB_USER"),
			os.Getenv("GITDM_GITHUB_OAUTH"),
			os.Getenv("GITDM_GITHUB_REPO"),
		),
	}
	env := map[string]string{"GIT_TERMINAL_PROMPT": "0"}
	execCommand(cmd, env, 0, []int{})
	fmt.Printf("get wd\n")
	wd, err := os.Getwd()
	if fatalOnError(err, false) {
		return
	}
	fmt.Printf("chdir gitdm\n")
	if fatalOnError(os.Chdir("gitdm"), false) {
		return
	}
	defer func() {
		fmt.Printf("chdir back to %s\n", wd)
		_ = os.Chdir(wd)
	}()
	fmt.Printf("sync repo\n")
	if !syncRepo() {
		return
	}
	w.WriteHeader(http.StatusOK)
	_, _ = io.WriteString(w, "SYNC_OK")
}

func handlePR(w http.ResponseWriter, req *http.Request) {
	info := requestInfo(req)
	fmt.Printf("Request: %s\n", info)
	var err error
	defer func() {
		fmt.Printf("Request(exit): %s err:%v\n", info, err)
	}()
	fmt.Printf("lock mutex\n")
	gMtx.Lock()
	defer func() {
		fmt.Printf("unlock mutex\n")
		gMtx.Unlock()
	}()
	gw = w
	path := html.EscapeString(req.URL.Path)
	// /pr/refs/pull/1/merge
	ary := strings.Split(path, "/")
	if len(ary) != 6 {
		fatalf(false, "malformed path:%s", path)
		return
	}
	prNumber, err := strconv.ParseInt(strings.TrimSpace(ary[4]), 10, 64)
	if err != nil {
		fatalf(false, "no PR number specified in path:%s:%v", path, err)
		return
	}
	fmt.Printf("checking PR %d\n", prNumber)
	fmt.Printf("Cleanup repo before\n")
	execCommand([]string{"rm", "-rf", "gitdm"}, nil, 1, []int{})
	defer func() {
		fmt.Printf("Cleanup repo after\n")
		execCommand([]string{"rm", "-rf", "gitdm"}, nil, 1, []int{})
	}()
	fmt.Printf("git clone\n")
	cmd := []string{
		"git",
		"clone",
		fmt.Sprintf(
			"https://%s:%s@github.com/%s",
			os.Getenv("GITDM_GITHUB_USER"),
			os.Getenv("GITDM_GITHUB_OAUTH"),
			os.Getenv("GITDM_GITHUB_REPO"),
		),
	}
	env := map[string]string{"GIT_TERMINAL_PROMPT": "0"}
	execCommand(cmd, env, 0, []int{})
	fmt.Printf("get wd\n")
	wd, err := os.Getwd()
	if fatalOnError(err, false) {
		return
	}
	fmt.Printf("chdir gitdm\n")
	if fatalOnError(os.Chdir("gitdm"), false) {
		return
	}
	defer func() {
		fmt.Printf("chdir back to %s\n", wd)
		_ = os.Chdir(wd)
	}()
	fmt.Printf("git fetch origin\n")
	_, ok := execCommand([]string{"git", "fetch", "origin", fmt.Sprintf("pull/%d/head:gitdm-sync-%d", prNumber, prNumber)}, nil, 1, []int{})
	if !ok {
		return
	}
	fmt.Printf("git checkout\n")
	_, ok = execCommand([]string{"git", "checkout", fmt.Sprintf("gitdm-sync-%d", prNumber)}, nil, 1, []int{})
	if !ok {
		return
	}
	defer func() {
		_, _ = execCommand([]string{"git", "checkout", "master"}, nil, 1, []int{})
	}()
	fmt.Printf("check repo PR %d\n", prNumber)
	if !checkRepo() {
		return
	}
	w.WriteHeader(http.StatusOK)
	_, _ = io.WriteString(w, "CHECK_OK")
}

func checkEnv() {
	requiredEnv := []string{
		"DA_API_URL",
		"GITDM_GITHUB_REPO",
		"GITDM_GITHUB_USER",
		"GITDM_GITHUB_OAUTH",
		"GITDM_GIT_USER",
		"GITDM_GIT_EMAIL",
	}
	for _, env := range requiredEnv {
		if os.Getenv(env) == "" {
			fatalf(true, "%s env variable must be set", env)
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
	http.HandleFunc("/push", handlePush)
	http.HandleFunc("/pr/", handlePR)
	fatalOnError(http.ListenAndServe("0.0.0.0:7070", nil), true)
}

func main() {
	serve()
	fatalf(true, "serve exited without error, returning error state anyway")
}
