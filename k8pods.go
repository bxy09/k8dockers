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
	Docker *K8PodWithDocker
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

func K8Generates(pods []K8Pod, dockerPods []K8PodWithDocker) (map[string][]K8Pod, []K8PodWithDocker) {
	remainK8PDocker := make([]K8PodWithDocker, 0)
	generates := make(map[string][]K8Pod)
	podsNameMap := make(map[string]int)
	for i := range pods {
		podsNameMap[pods[i].Metadata.Namespace+"/"+pods[i].Metadata.Name] = i
	}
	for i := range dockerPods {
		name := dockerPods[i].Name
		idx, exist := podsNameMap[name]
		if !exist {
			remainK8PDocker = append(remainK8PDocker, dockerPods[i])
			continue
		}
		pods[idx].Docker = &dockerPods[i]
	}
	for i := range pods {
		gName := pods[i].Metadata.GenerateName
		generates[gName] = append(generates[gName], pods[i])
	}
	return generates, remainK8PDocker
}
