package main

import (
	// "fmt"
	// "io/ioutil"
	"net/http"
	"os/exec"
	"sync"
)

var (
	mutex sync.Mutex
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		username, password, ok := r.BasicAuth()
        if !ok {
            w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)
            http.Error(w, "Unauthorized", http.StatusUnauthorized)
            return
        }

        if username != "admin" || password != "password" {
            http.Error(w, "Unauthorized", http.StatusUnauthorized)
            return
        }

		if r.URL.Path == "/" {
			http.ServeFile(w, r, "./index.html")
			return
		}

		if r.URL.Path == "/get" {
			if mutex.TryLock() {
				defer mutex.Unlock()
				cmd := exec.Command("./exe.sh")
				stdout, err := cmd.Output()
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
				w.Write(stdout)
				return
			} else {
				w.WriteHeader(http.StatusServiceUnavailable)
				return
			}
		}

		http.NotFound(w, r)
	})

	http.ListenAndServe(":8080", nil)
}