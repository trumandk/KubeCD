package main

import (
	"fmt"
	"github.com/go-git/go-git"
	"github.com/go-git/go-git/plumbing/object"
	"net/http"
)

func menu(w http.ResponseWriter, req *http.Request) {
        fmt.Fprintf(w, "<head>")
        fmt.Fprintf(w, "<title>KubeCD</title>")
//        fmt.Fprintf(w, " <link rel=\"icon\" type=\"image/png\" href=\"files/jumpstarter.png\">")
        fmt.Fprintf(w, "</head>")
        fmt.Fprintf(w, "<link rel=\"stylesheet\" href=\"files/bootstrap.css\">")
        fmt.Fprintf(w, "<script src=\"files/bootstrap.js\"></script>")
        fmt.Fprintf(w, "<body>")
        fmt.Fprintf(w, "<nav class=\"navbar navbar-expand-lg navbar-dark bg-dark\">")
        fmt.Fprintf(w, "<a class=\"navbar-brand\" href=\"/pods\">Pods</a>")
        fmt.Fprintf(w, "<a class=\"navbar-brand\" href=\"/nodes\">Nodes</a>")
        fmt.Fprintf(w, "<a class=\"navbar-brand\" href=\"/services\">Services</a>")
        fmt.Fprintf(w, "<a class=\"navbar-brand\" href=\"/configmaps\">Configmaps</a>")
        fmt.Fprintf(w, "<a class=\"navbar-brand\" href=\"/namespaces\">Namespaces</a>")
        fmt.Fprintf(w, "<a class=\"navbar-brand\" href=\"/apply\">Apply</a>")
        fmt.Fprintf(w, "<a class=\"navbar-brand\" href=\"/git\">Git</a>")
        fmt.Fprintf(w, "</nav>")

}


func gitWeb(w http.ResponseWriter, req *http.Request) {
	menu(w, req)
	r, err := git.PlainOpen("/git/")
	if err != nil {
		fmt.Printf("plain open :%s", err)
	}
	/*
		work, err := r.Worktree()
		if err != nil {
			fmt.Printf("worktree error :%s", err)
		}*/
	ref, err := r.Head()
	if err != nil {
		fmt.Printf("head error :%s", err)
	}
	cIter, err := r.Log(&git.LogOptions{From: ref.Hash()})
	if err != nil {
		fmt.Printf("log error :%s", err)
	}

	fmt.Fprintf(w, "<pre><code>\n")
	err = cIter.ForEach(func(c *object.Commit) error {
		fmt.Fprintf(w, "%s\n", c.String())

		return nil
	})
	fmt.Fprintf(w, "</code></pre>\n")

	if err != nil {
		fmt.Printf("cIter.ForEach error :%s", err)
	}
}
