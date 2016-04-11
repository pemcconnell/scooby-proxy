package main

import (
	"flag"
	"log"
	"os"
	"text/template"
)

type NginxConf struct {
	Resolver         string
	Subdomain        string
	Tld              string
	Port             string
	HtpasswdEnabled  bool
}

func main() {
	
	resolver := flag.String("resolver", "", "Specify the resolver")
	subdomain := flag.String("subdomain", "", "Specify the subdomain")
	tld := flag.String("tld", "", "Specify the top-level domain")
	port := flag.String("port", "80", "Specify the port")
	htpasswd := flag.String("htpasswd", "scooby:$apr1$RCZ6WLU9$AM8ha8UfdrxmiPOG.wqjf0", "Enter the htpasswd") // default: scooby/doo
	flag.Parse()
	conf := NginxConf{
		Resolver:         *resolver,
		Subdomain:        *subdomain,
		Tld:              *tld,
		Port:             *port,
		HtpasswdEnabled:  true,
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
	f, err := os.Create("/etc/nginx/conf.d/" + *subdomain)
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
	data := []byte(*htpasswd)
	f, err = os.Create(dir + *subdomain)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	_, err = f.Write(data)
	if err != nil {
		log.Fatal(err)
	}

}
