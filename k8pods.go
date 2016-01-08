package k8dockers

import (
	"github.com/fsouza/go-dockerclient"
)

type K8PodWithDocker struct {
	Name      string
	Container []docker.APIContainers `json:",omitempty"`
}

type K8Pod struct {
	Metadata struct {
		Name         string
		GenerateName string
		Namespace    string
	}
	Spec   struct{}
	Status struct {
		Status string
	}
}

func ReadK8PodsFrom(containers []docker.APIContainers) ([]K8PodWithDocker, []docker.APIContainers) {
	pods := make([]K8PodWithDocker, 0, len(containers)/2)
	remains := make([]docker.APIContainers, 0)
	podKeyMapping := make(map[string]int)
	for i := range containers {
		podName := containers[i].Labels["io.kubernetes.pod.name"]
		if podName == "" {
			remains = append(remains, containers[i])
			continue
		}
		idx, ok := podKeyMapping[podName]
		if !ok {
			idx = len(pods)
			pods = append(pods, K8PodWithDocker{Name: podName})
			podKeyMapping[podName] = idx
		}
		pods[idx].Container = append(pods[idx].Container, containers[i])
	}
	return pods, remains
}

