package main

import (
	"fmt"
	"os"
	"github.com/go-git/go-git"
	"github.com/go-git/go-git/plumbing/object"
	"github.com/go-git/go-git/plumbing/transport"
	"github.com/go-git/go-git/plumbing/transport/ssh"
	"time"
        "log"
	"net/http"
	"os/exec"
	"crypto/subtle"
	"bufio"
)

var username = os.Getenv("JUMPSTARTER_USERNAME")
var password = os.Getenv("JUMPSTARTER_PASSWORD")

func BasicAuth(handler http.HandlerFunc) http.HandlerFunc {
        realm := "Please enter your username and password for this site"
        return func(w http.ResponseWriter, r *http.Request) {

                user, pass, ok := r.BasicAuth()

                if !ok || subtle.ConstantTimeCompare([]byte(user), []byte(username)) != 1 || subtle.ConstantTimeCompare([]byte(pass), []byte(password)) != 1 {
                        w.Header().Set("WWW-Authenticate", `Basic realm="`+realm+`"`)
                        w.WriteHeader(401)
                        w.Write([]byte("Unauthorised.\n"))
                        return
                }
                handler(w, r)
        }
}

func MyPublicKeys() transport.AuthMethod {
	publicKeys, err := ssh.NewPublicKeysFromFile("git", "/root/.ssh/id_rsa", "")
	if err != nil {
		panic(err)
	}
	return publicKeys
}

var publicKeys = MyPublicKeys()

func dockerInitGit() {
	err := os.RemoveAll("/git/")
	if err != nil {
		fmt.Printf("Remove folder :%s", err)
	}

	r, err := git.PlainClone("/git/", false, &git.CloneOptions{
		URL:      os.Getenv("GIT_CLUSTER"),
		Auth:     publicKeys,
		Progress: os.Stdout,
	})
	if err != nil {
		fmt.Printf("Git error :%s", err)
	}
	ref, err := r.Head()
	fmt.Printf("ref :%s", ref)
	if err != nil {
		fmt.Printf("head :%s", err)
	}
	fmt.Printf("ref :%s", ref)

}
func dockerGitCommit(path, hash string) {
	kubeCD_log := path +"KubeCD.log"
	file, err := os.OpenFile(kubeCD_log, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	datawriter := bufio.NewWriter(file)

	datawriter.WriteString("#### Updated hash:" + hash + " ########### " + time.Now().String() + " #####\n")
	datawriter.WriteString(kubectlStatus("/kubectl","get", "nodes,pods,services,configmaps", "--all-namespaces", "-o", "wide"))

	datawriter.Flush()
	file.Close()

	if err != nil {
		log.Fatalf("failed creating file: %s", err)
	}
	r, err := git.PlainOpen(path)
	if err != nil {
		fmt.Printf("plain open :%s", err)
	}
	w, err := r.Worktree()
	if err != nil {
		fmt.Printf("worktree error :%s", err)
	}
	w.Add("KubeCD.log")

	commit, err2 := w.Commit("Kubernetes log:KubeCD.log", &git.CommitOptions{
		Author: &object.Signature{
			Name:  "KubeCD",
			Email: "auto@kubecd.io",
			When:  time.Now(),
		},
	})
	if err2 != nil {
		fmt.Printf("Commit error:%s", err2)
	}
	obj, err3 := r.CommitObject(commit)
	if err3 != nil {
		fmt.Printf("CommitObject :%s", err3)
	}
	fmt.Println(obj)

	err5 := r.Push(&git.PushOptions{
		RemoteName: "origin",
		Auth:       publicKeys,
		Progress:   os.Stdout,
	})
	if err5 != nil {
		fmt.Printf("push :%s", err5)
	}
	ref, err := r.Head()
	fmt.Printf("ref :%s", ref)
	if err != nil {
		fmt.Printf("head :%s", err)
	}
}

func dockerGitUpdate(path string) {
	r, err := git.PlainOpen(path)
	if err != nil {
		fmt.Printf("plain open :%s", err)
		return
	}
	w, err := r.Worktree()
	if err != nil {
		fmt.Printf("worktree error :%s", err)
		return
	}

	err = w.Pull(&git.PullOptions{
		RemoteName: "origin",
		Auth:       publicKeys,
		Progress:   os.Stdout,
	})
	if err != nil && err != git.NoErrAlreadyUpToDate {
		fmt.Printf("pull error :%s\n", err)
		return
	}
	if err == git.NoErrAlreadyUpToDate {
		return
	}
	fmt.Printf("pull new files \n")
	ref, _ := r.Head()
		fmt.Println(kubectlCommand(path))
		time.Sleep(10 * time.Second)
		dockerGitCommit(path, ref.Hash().String())
}

func kubectlCommand(path string) string {
	out, err := exec.Command("/kubectl", "apply","--prune", "-f", path,"--recursive", "--all", "--wait").CombinedOutput()
	if err != nil {
		fmt.Printf("Error updating:%s Message:%s", path, err)
	}
	output := string(out[:])
	return output
}

func CommandWeb(name string, arg ...string) http.HandlerFunc {
        return func(w http.ResponseWriter, r *http.Request) {
	menu(w, r)
	fmt.Fprintf(w, "<pre>")
        fmt.Fprintf(w, kubectlStatus(name, arg...))
	fmt.Fprintf(w, "</pre>")
        }
}

func kubectlStatus(name string, arg ...string) string {
	out, err := exec.Command(name, arg...).CombinedOutput()

	if err != nil {
		fmt.Printf("Error get status Message:%s", err)
	}
	output := string(out[:])
	return output
}
func ApplyKube(path string) http.HandlerFunc {
        return func(w http.ResponseWriter, r *http.Request) {
	menu(w, r)
	fmt.Fprintf(w, "<pre>")
        fmt.Fprintf(w, kubectlCommand(path))
	fmt.Fprintf(w, "</pre>")
        }
}

func main() {

	dockerInitGit()
	fmt.Println(kubectlCommand("/git/"))

	go func() {
		for {
			dockerGitUpdate("/git/")
			fmt.Println(kubectlCommand("/git/"))
			time.Sleep(1 * time.Second)
		}
	}()

	fileServer := http.FileServer(http.Dir("/files"))
//	gitFileServer := http.FileServer(http.Dir("/git/"))
        mux := http.NewServeMux()
	mux.Handle("/files/", http.StripPrefix("/files", fileServer))
//	mux.Handle("/gitfiles/", BasicAuth(http.StripPrefix("/gitfiles", gitFileServer)))

        mux.HandleFunc("/",		BasicAuth(CommandWeb("/kubectl","get", "pods", "--all-namespaces", "-o", "wide")))
        mux.HandleFunc("/deployment", 	BasicAuth(CommandWeb("/kubectl","get", "deployment", "--all-namespaces", "-o", "wide")))
        mux.HandleFunc("/pv", 		BasicAuth(CommandWeb("/kubectl","get", "pv", "--all-namespaces", "-o", "wide")))
        mux.HandleFunc("/pods", 	BasicAuth(CommandWeb("/kubectl","get", "pods", "--all-namespaces", "-o", "wide")))
        mux.HandleFunc("/nodes",	BasicAuth(CommandWeb("/kubectl","get", "nodes", "--all-namespaces", "-o", "wide")))
        mux.HandleFunc("/services",	BasicAuth(CommandWeb("/kubectl","get", "services", "--all-namespaces", "-o", "wide")))
        mux.HandleFunc("/configmaps",	BasicAuth(CommandWeb("/kubectl","get", "configmaps", "--all-namespaces", "-o", "wide")))
        mux.HandleFunc("/namespaces",	BasicAuth(CommandWeb("/kubectl","get", "namespaces", "--all-namespaces", "-o", "wide")))
        mux.HandleFunc("/toppod",	BasicAuth(CommandWeb("/kubectl","top", "pod")))
        mux.HandleFunc("/topnode",	BasicAuth(CommandWeb("/kubectl","top", "node")))
        mux.HandleFunc("/apply",	BasicAuth(ApplyKube("/git/")))
        mux.HandleFunc("/git",		BasicAuth(gitWeb))
	log.Println("Starting server on :8042")
        log.Fatal(http.ListenAndServe(":8042", mux))
}
