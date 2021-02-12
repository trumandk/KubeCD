package main

import (
	"fmt"
        "log"
	"net/http"
	"os/exec"
)


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
	fmt.Fprintf(w, "<p><b>")
	fmt.Fprintf(w, name)
	for _, n := range arg {
		fmt.Fprintf(w, " " + n)
	}
	fmt.Fprintf(w, "</b></p>")
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

func handleFileServer(dir, prefix string) http.HandlerFunc {
    fs := http.FileServer(http.Dir(dir))
    realHandler := http.StripPrefix(prefix, fs).ServeHTTP
    return func(w http.ResponseWriter, req *http.Request) {
        realHandler(w, req)
    }
}

func main() {

        mux := http.NewServeMux()

	mux.HandleFunc("/files/", handleFileServer("/files/", "/files"))

        mux.HandleFunc("/",		CommandWeb("/kubectl","get", "pods", "--all-namespaces", "-o", "wide"))
        mux.HandleFunc("/deployment", 	CommandWeb("/kubectl","get", "deployment", "--all-namespaces", "-o", "wide"))
        mux.HandleFunc("/pv", 		CommandWeb("/kubectl","get", "pv", "--all-namespaces", "-o", "wide"))
        mux.HandleFunc("/pods", 	CommandWeb("/kubectl","get", "pods", "--all-namespaces", "-o", "wide"))
        mux.HandleFunc("/nodes",	CommandWeb("/kubectl","get", "nodes", "--all-namespaces", "-o", "wide"))
        mux.HandleFunc("/services",	CommandWeb("/kubectl","get", "services", "--all-namespaces", "-o", "wide"))
        mux.HandleFunc("/configmaps",	CommandWeb("/kubectl","get", "configmaps", "--all-namespaces", "-o", "wide"))
        mux.HandleFunc("/namespaces",	CommandWeb("/kubectl","get", "namespaces", "--all-namespaces", "-o", "wide"))
        mux.HandleFunc("/toppod",	CommandWeb("/kubectl","top", "pod"))
        mux.HandleFunc("/topnode",	CommandWeb("/kubectl","top", "node"))
        mux.HandleFunc("/events",	CommandWeb("/kubectl", "get", "events", "--sort-by=.metadata.creationTimestamp", "--all-namespaces"))
	log.Println("Starting server on :8080")
	
        log.Fatal(http.ListenAndServe(":8080", mux))
}
