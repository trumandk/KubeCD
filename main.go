package main

import (
	"fmt"
	"os"
	"github.com/go-git/go-git"
	"github.com/go-git/go-git/plumbing/object"
	"github.com/go-git/go-git/plumbing/transport"
	"github.com/go-git/go-git/plumbing/transport/ssh"
	"time"
	 "path/filepath"
        "log"
	"os/exec"
	"bufio"
)

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
	datawriter.WriteString(kubectlStatus())

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

	commit, err2 := w.Commit("Auto-commit server:"+kubeCD_log, &git.CommitOptions{
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
	head, _ := r.Head()
	headCommit, _ := r.CommitObject(head.Hash())
	headTree, _ := headCommit.Tree()
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
	commit, _ := r.CommitObject(ref.Hash())

	fmt.Println(commit)
	tree, _ := commit.Tree()

	changes, err := headTree.Diff(tree)
	if err != nil {
		fmt.Printf("diff error :%s\n", err)
		return
	}
		for _, c := range changes {
			action, _ := c.Action()
			fmt.Printf("changes to:%s from:%s what:%s\n", c.To.Name,c.From.Name, action.String())
			if action.String() == "Insert" {
				fmt.Printf("Insert to:%s\n", c.To.Name)
				kubectlCommand("apply", path + c.To.Name)
			}
			if action.String() == "Delete" {
				fmt.Printf("Delete from:%s\n", c.From.Name)
				kubectlCommand("delete", path + c.From.Name)
			}
			if action.String() == "Modify" {
				fmt.Printf("Modify to:%s\n", c.To.Name)
				kubectlCommand("apply", path + c.To.Name)
			}
		}
		time.Sleep(10 * time.Second)
		dockerGitCommit(path, ref.Hash().String())
}

func kubectlCommand(what string, file string) {
	out, err := exec.Command("/kubectl", what, "-f", file, "--wait").CombinedOutput()

	if err != nil {
		fmt.Printf("Error updating:%s Message:%s", file, err)
	}
	output := string(out[:])
	fmt.Println(output)
}

func kubectlStatus() string {
	out, err := exec.Command("/kubectl","get", "pods", "--all-namespaces").CombinedOutput()

	if err != nil {
		fmt.Printf("Error get status Message:%s", err)
	}
	output := string(out[:])
	fmt.Println(output)
	return output
}

func kubeInit(path string) {
	filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
        if err != nil {
            log.Fatalf(err.Error())
        }
	if filepath.Ext(info.Name()) == ".yaml" {
		fmt.Printf("File Name: %s\n", path)
		kubectlCommand("apply", path)
	}
        return nil
    })
}

func main() {
	if _, err := os.Stat("/git/.git/"); err == nil {
		fmt.Printf("Using existing git repo \n")
	} else if os.IsNotExist(err) {
		dockerInitGit()
		kubeInit("/git/")
	}
	for {
		dockerGitUpdate("/git/")
	//	kubectlStatus()
		time.Sleep(1 * time.Second)
	}
}
