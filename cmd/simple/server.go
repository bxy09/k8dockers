package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/Sirupsen/logrus"
	"github.com/bxy09/k8dockers"
	"github.com/fsouza/go-dockerclient"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"net/http"
)

type inputConfig struct {
	Dockers []string
	K8API   string
	Modules []struct {
		Name string
		Min  int
	}
}

func main() {
	var configPath string
	flag.StringVar(&configPath, "f", "", "Config file")
	flag.Parse()
	if configPath == "" {
		fmt.Println("[Usage] server -f [Yaml file]")
		flag.PrintDefaults()
		return
	}
	logrus.SetFormatter(&logrus.JSONFormatter{})
	bytes, err := ioutil.ReadFile(configPath)
	if err != nil {
		logrus.Fatal("Cannot read the file, ", err.Error())
	}
	var config inputConfig
	err = yaml.Unmarshal(bytes, &config)
	if err != nil {
		logrus.Fatal("Cannot parse the yaml file, ", err.Error())
	}
	if config.K8API == "" {
		logrus.Fatal("K8API cannot be empty")
	}
	clients := make([]*docker.Client, len(config.Dockers))
	for i := range clients {
		logrus.Info("Load from ", config.Dockers[i])
		clients[i], err = docker.NewClient(config.Dockers[i])
		if err != nil {
			logrus.Fatal("Cannot connet to client,", config.Dockers[i], ", ", err.Error())
		}
	}
	http.HandleFunc("/dockers/pods/json", func(w http.ResponseWriter, r *http.Request) {
		logrus.WithFields(logrus.Fields{
			"from": r.RemoteAddr,
		}).Info("Request /dockers/pods/json")
		type podsJson struct {
			Node  string
			Error string `json:",omitempty"`
			Pods  []k8dockers.K8PodWithDocker
		}
		result := make([]podsJson, len(clients))
		for i := range clients {
			result[i].Node = config.Dockers[i]
			containers, err := clients[i].ListContainers(docker.ListContainersOptions{All: false})
			if err != nil {
				result[i].Error = err.Error()
				continue
			}
			pods, remains := k8dockers.ReadK8PodsFrom(containers)
			if len(remains) != 0 {
				result[i].Error = fmt.Sprintf("there are %d unknown containers", len(remains))
			}
			result[i].Pods = pods
		}
		bytes, err := json.Marshal(result)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			logrus.Error("Cannot unmarshal result, ", err.Error())
			return
		}
		w.Write(bytes)
		return
	})
	k8ApiPodsJson := fmt.Sprintf("http://%s/api/v1/pods", config.K8API)
	http.HandleFunc("/k8/pods/json", func(w http.ResponseWriter, r *http.Request) {
		logrus.WithFields(logrus.Fields{
			"from": r.RemoteAddr,
		}).Info("Request /k8/pods/json")
		response, err := http.Get(k8ApiPodsJson)
		if err != nil {
			logrus.Error("Cannot visit the k8 server, ", err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		defer response.Body.Close()
		body, err := ioutil.ReadAll(response.Body)
		if err != nil {
			logrus.Error("Cannot read all from k8: ", err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		_, err = w.Write(body)
		if err != nil {
			logrus.Error("Cannot rewrite k8 result, ", err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		return
	})

	http.HandleFunc("/modules/expect/json", func(w http.ResponseWriter, r *http.Request) {
		logrus.WithFields(logrus.Fields{
			"from": r.RemoteAddr,
		}).Info("Request /modules/expect/json")
		bytes, err := json.Marshal(config.Modules)
		if err != nil {
			logrus.Error("Cannot marshal data for modules/expect, ", err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Write(bytes)
		return
	})

	err = http.ListenAndServe("127.0.0.1:8011", nil)
	if err != nil {
		logrus.Info("service quit as:", err.Error())
	}
	return
}
