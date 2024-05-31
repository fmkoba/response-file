package main

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type Route struct {
	Path string
	File string
}

func (r Route) GetRoute() string {
	result := strings.Replace(r.Path, "routes/", "/", 1)
	resultSlice := strings.Split(result, ".")
	return resultSlice[0]
}

func getResponse(path string, routes []*Route) (route Route, err error) {
	for _, r := range routes {
		if r.GetRoute() == path {
			return *r, nil
		}
	}
	return Route{}, errors.New("404: Not found")
}

func main() {
	config := readConf()
	fmt.Println(config)

	var routes = []*Route{}

	filepath.WalkDir("./routes", func(path string, dir fs.DirEntry, err error) error {
		if !dir.IsDir() {
			routes = append(routes, &Route{Path: path, File: dir.Name()})
		}
		return nil
	})
	fmt.Println("----- Routes ------")
	for _, v := range routes {
		fmt.Println("http://localhost:" + config.Port + v.GetRoute())
	}
	fmt.Println("-------------------")

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		route, err := getResponse(r.URL.Path, routes)
		if err != nil {
			w.WriteHeader(404)
			io.WriteString(w, "Not found!\n")
		} else {
			fmt.Println("[" + r.Method + "] " + r.URL.Path)
			w.Header().Add("content-type", "application/json")

			io.WriteString(w, readJSONfromFile(route.Path))
		}
	})

	http.ListenAndServe(":"+config.Port, nil)
}

func readJSONfromFile(path string) string {
	content, err := os.ReadFile(path)
	if err != nil {
		panic(err)
	}
	return string(content)
}

type Config struct {
	Port     string
	ShowLogs bool
}

func readConf() Config {
	file, err := os.ReadFile("config.yml")
	config := Config{Port: "8080", ShowLogs: true}
	if err != nil {
		return config
	}

	portRegex := regexp.MustCompile("^[pP]ort: ([0-9]*)")
	portMatch := portRegex.FindStringSubmatch(string(file))
	if len(portMatch) == 2 {
		config.Port = portMatch[1]
	}
	return config
}
