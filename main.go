package main

import (
	"log"
	"os"
	"net/http"
	"os/exec"
	"fmt"
	"encoding/json"
	"io/ioutil"
	"text/template"
	"crypto/tls"
)

type ApiResponsePods struct {
	Items []ApiResponsePodItems `json:"items"`
}

type ApiResponsePodItems struct {
	Metadata ApiResponsePodItemMetadata `json:"metadata"`
	Spec     ApiResponsePodItemSpec `json:"spec"`
}

type ApiResponsePodItemMetadata struct {
	Name string `json:"name"`
	Labels ApiResponsePodItemMetadataLabel `json:"labels"`
}

type ApiResponsePodItemMetadataLabel struct {
	Name string `json:"name"`
}

type ApiResponsePodItemSpec struct {
	Containers []ApiResponsePodItemSpecContainers `json:"containers"`
}

type ApiResponsePodItemSpecContainers struct {
	Name string `json:"name"`
	Image string `json:"image"`
	Ports []ApiResponsePodItemSpecContainerPorts `json:"ports"`
}

type ApiResponsePodItemSpecContainerPorts struct {
	ContainerPort int `json:"containerPort"`
}

type NginxConf struct {
	Resolver         string
	Subdomain        string
	Tld              string
	Port             int
	HtpasswdEnabled  bool
}

type Config struct {
	Resolver string `json:"resolver"`
	Tld string `json:"tld"`
	HtpasswdEnabled bool `json:"htpasswdenabled"`
	DefaultHtpasswd string `json:"defaulthtpasswd"`
	Namespace string `json:"namespace"`
}

func main() {
	// load config
	url := "config.json"
	file, e := ioutil.ReadFile(url)
	if e != nil {
        fmt.Printf("File error: %v\n", e)
        os.Exit(1)
    }
    var config Config
    json.Unmarshal(file, &config)
    

	generateConfigsFromKubernetesAPI(config)

	// reload nginx
	cmd := "nginx"
	args := []string{"-s", "reload"}
	if err := exec.Command(cmd, args...).Run(); err != nil {
		os.Exit(1)
	}
}

func generateConfigsFromKubernetesAPI(config Config) {
	token, err := ioutil.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/token")
	if err != nil {
		log.Fatal(err)
	}
	tr := &http.Transport{
        TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
    }
    client := &http.Client{Transport: tr}
	url := "https://" + os.Getenv("KUBERNETES_SERVICE_HOST") + ":" + os.Getenv("KUBERNETES_PORT_443_TCP_PORT") + "/api/v1/namespaces/" + config.Namespace + "/pods"
	req, err := http.NewRequest("GET", url, nil)
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", string(token)))
	res, err := client.Do(req)
    if err != nil {
    	log.Fatal(err)
    }
    defer res.Body.Close()

	file, err := ioutil.ReadAll(res.Body)
	if err != nil {
        log.Fatal(err)
    }
    var pods ApiResponsePods
    json.Unmarshal(file, &pods)
    for i := 0; i < len(pods.Items); i++ {
    	createNginxConfig(config, pods.Items[i].Metadata.Labels.Name, pods.Items[i].Spec.Containers[0].Ports[0].ContainerPort)
    }
}

func createNginxConfig(config Config, subdomain string, port int) {
	htpasswd := config.DefaultHtpasswd
	conf := NginxConf{
		Resolver:         config.Resolver,
		Tld:              config.Tld,
		HtpasswdEnabled:  config.HtpasswdEnabled,
		Subdomain:        subdomain,
		Port:             port,
	}
	nginxTemplate, err := template.New("nginxconf").Parse(`server {
    listen       80;
    server_name  {{.Subdomain}}.{{.Tld}};

    error_log /var/log/nginx/error.log;
    access_log /var/log/nginx/access.log;

    location / {
        {{if .HtpasswdEnabled}}
        auth_basic                       "Restricted";
        auth_basic_user_file             /etc/nginx/htpasswd.d/{{.Subdomain}};
        {{end}}
        resolver            {{.Resolver}};
        proxy_pass          http://{{.Subdomain}}.default.svc.cluster.local:{{.Port}}; 
        proxy_set_header    X-Real-IP   $remote_addr;
        proxy_set_header    Upgrade     $http_upgrade;
        proxy_set_header    Host        $http_host;
    }
}`)
	if err != nil {
		log.Fatal(err)
	}
	f, err := os.Create("/etc/nginx/conf.d/" + subdomain)
	if err != nil {
		log.Println("create file: ", err)
		return
	}
	err = nginxTemplate.Execute(f, conf)
	if err != nil {
		log.Fatal(err)
	}
	f.Close()
	// htpasswd
	dir := "/etc/nginx/htpasswd.d/"
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err = os.Mkdir(dir, 0755)
	}
	if err != nil {
		log.Fatal(err)
	}
	data := []byte(htpasswd)
	f, err = os.Create(dir + subdomain)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	_, err = f.Write(data)
	if err != nil {
		log.Fatal(err)
	}

}
