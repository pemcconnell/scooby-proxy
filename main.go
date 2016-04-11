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
	nginxTemplate, err := template.ParseFiles("nginxconf.tmpl")
	if err != nil {
		log.Fatal(err)
	}
	err = nginxTemplate.Execute(os.Stdout, conf)
	if err != nil {
		log.Fatal(err)
	}
	
	// htpasswd
	dir := "/etc/nginx/htpasswd.d/"
	err = os.Mkdir(dir, 0755)
	if err != nil {
		log.Fatal(err)
	}
	data := []byte(*htpasswd)
	f, err := os.Create(dir + *subdomain)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	_, err = f.Write(data)
	if err != nil {
		log.Fatal(err)
	}

}
