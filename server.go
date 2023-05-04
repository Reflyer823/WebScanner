package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os/exec"
	"sync"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Port int `yaml:"port"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	EnableHTTPS bool `yaml:"enableHttps"`
}

var (
	mutex sync.Mutex
	config Config
)

func main() {
	// Load Config File
	config_str, err := ioutil.ReadFile("./config.yaml")
	if err != nil {
		fmt.Println(err)
		return
	}

	err = yaml.Unmarshal(config_str, &config)
	if err != nil {
		fmt.Println(err)
		return
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		username, password, ok := r.BasicAuth()
        if !ok {
            w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)
            http.Error(w, "Unauthorized", http.StatusUnauthorized)
            return
        }

        if username != config.Username || password != config.Password {
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

	addr := fmt.Sprintf(":%d", config.Port)
	if config.EnableHTTPS {
		err = http.ListenAndServeTLS(addr, "cert.pem", "key.pem", nil)
		if err != nil {
			fmt.Println(err)
			return
		}
	} else {
		err = http.ListenAndServe(addr, nil)
		if err != nil {
			fmt.Println(err)
			return
		}
	}
}