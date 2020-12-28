package main

import (
	"fmt"
	"os"
	"github.com/go-git/go-git"
	//"github.com/go-git/go-git/utils/merkletrie"
	//"github.com/go-git/go-git/utils/merkletrie/noder"
	"github.com/go-git/go-git/plumbing/object"
	"github.com/go-git/go-git/plumbing/transport"
	"github.com/go-git/go-git/plumbing/transport/ssh"
	"time"
	//"io"
	"io/ioutil"
	"os/exec"
	"log"
	//"bytes"
	/*
	"bytes"
	"crypto/subtle"
	"github.com/tidwall/sjson"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	*/
)

func MyPublicKeys() transport.AuthMethod {
	publicKeys, err := ssh.NewPublicKeysFromFile("git", "/root/.ssh/id_rsa", "")

	if err != nil {
		panic(err)
	}
	return publicKeys
}

var publicKeys = MyPublicKeys()

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}



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
func dockerGitCommit(filename string) {
	r, err := git.PlainOpen("/git/")
	if err != nil {
		fmt.Printf("plain open :%s", err)
	}
	w, err := r.Worktree()
	if err != nil {
		fmt.Printf("worktree error :%s", err)
	}
	w.Add(filename)

	commit, err2 := w.Commit("Auto-commit server:"+filename, &git.CommitOptions{
		Author: &object.Signature{
			Name:  "JumpStarter",
			Email: "jumpstarter@jumpstarter.io",
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

func dockerGitUpdate() {
	r, err := git.PlainOpen("/git/")
	head, _ := r.Head()
//	hash := head.Hash()
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
	//headTree, _ := HeadCommit.Tree()

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
			}
			if action.String() == "Delete" {
				fmt.Printf("Delete from:%s\n", c.From.Name)
			}
			if action.String() == "Modify" {
				fmt.Printf("Modify to:%s\n", c.To.Name)
			}
		}
	//merkletrie.DiffTree(tree, tree, isEquals)
/*
	tree.Files().ForEach(func(f *object.File) error {
		fmt.Printf("100644 blob %s    %s\n", f.Hash, f.Name)
		return nil
	})*/
//	fileStats, _ := r.StatsContext(context.Background())
//	fileStats[0].Addition
}
/*
func isEquals(a, b noder.Hasher) bool {
	if bytes.Equal(a.Hash(), empty) || bytes.Equal(b.Hash(), empty) {
		return false
	}

	return bytes.Equal(a.Hash(), b.Hash())
}
*/
func dockerRun(ip string, file string) {
	out, err := exec.Command("/usr/bin/docker-compose", "--compatibility", "-p", file, "--env-file", "/git/docker/env", "-H", "ssh://core@"+ip, "-f", "/git/docker/"+file, "up", "-d", "--remove-orphans").CombinedOutput()

	if err != nil {
		fmt.Printf("Error updating:%s Message:%s", ip, err)
	}
	output := string(out[:])
	//	if len(output) > 0 {
	fmt.Println(output)
	//	}
}

func dockerClean(ip string, file string) {
	out, err := exec.Command("/usr/bin/docker-compose", "-p", file, "-H", "ssh://core@"+ip).CombinedOutput()

	if err != nil {
		fmt.Printf("Error updating:%s Message:%s", ip, err)
	}
	output := string(out[:])
	//	if len(output) > 0 {
	fmt.Println(output)
	//	}
}

func dockercompose() {

	nodes, err := ioutil.ReadDir("/git/docker/")
	if err != nil {
		log.Fatal(err)
	}
	for _, f := range nodes {
		if f.Name() != "env" {
			fmt.Printf("docker-compose :%s\n", f.Name())

//			dockerRun(f.Name(), "all")
//			dockerRun(f.Name(), f.Name())
		}
	}
}


func main() {
	dockerInitGit()
	for {
		dockerGitUpdate()
//		dockercompose()
		time.Sleep(1 * time.Second)
	}
}
