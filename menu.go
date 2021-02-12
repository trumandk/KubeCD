package main

import (
	"fmt"
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
        fmt.Fprintf(w, "<nav class=\"navbar navbar-expand-lg navbar-dark bg-primary\">")
        fmt.Fprintf(w, "<a class=\"navbar-brand\" href=\"/pods\">Pods</a>")
        fmt.Fprintf(w, "<a class=\"navbar-brand\" href=\"/deployment\">Deployment</a>")
        fmt.Fprintf(w, "<a class=\"navbar-brand\" href=\"/nodes\">Nodes</a>")
        fmt.Fprintf(w, "<a class=\"navbar-brand\" href=\"/services\">Services</a>")
        fmt.Fprintf(w, "<a class=\"navbar-brand\" href=\"/configmaps\">Configmaps</a>")
        fmt.Fprintf(w, "<a class=\"navbar-brand\" href=\"/namespaces\">Namespaces</a>")
        fmt.Fprintf(w, "<a class=\"navbar-brand\" href=\"/pv\">PersistentVolumes</a>")
        fmt.Fprintf(w, "<a class=\"navbar-brand\" href=\"/toppod\">TopPod</a>")
        fmt.Fprintf(w, "<a class=\"navbar-brand\" href=\"/topnode\">TopNode</a>")
        fmt.Fprintf(w, "<a class=\"navbar-brand\" href=\"/events\">Events</a>")
        fmt.Fprintf(w, "</nav>")

}

