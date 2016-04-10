package main

import (
	"flag"
	"log"
	"os"
	"text/template"
)

type NginxConf struct {
	Resolver  *string
	Subdomain *string
	Tld       *string
	Port      *string
	Htpasswd  bool
}

func main() {
	var (
		resolver  = flag.String("resolver", "", "Specify the resolver")
		subdomain = flag.String("subdomain", "", "Specify the subdomain")
		tld       = flag.String("tld", "", "Specify the top-level domain")
		port      = flag.String("port", "", "Specify the port")
	)
	conf := NginxConf{
		Resolver:  resolver,
		Subdomain: subdomain,
		Tld:       tld,
		Port:      port,
		Htpasswd:  true,
	}
	nginxTemplate, err := template.ParseFiles("nginxconf.tmpl")
	if err != nil {
		log.Fatal(err)
	}
	err = nginxTemplate.Execute(os.Stdout, conf)
	if err != nil {
		log.Fatal(err)
	}
}
