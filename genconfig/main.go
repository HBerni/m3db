package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"text/template"

	inserts "github.com/m3db/m3db/genconfig/inserts"
	yaml "gopkg.in/yaml.v2"
)

type BuildConfigFunc func(string, string) (string, error)

func main() {

	var inserts = map[string]BuildConfigFunc{
		"m3em_agent": buildM3EmAgentConfig,
		"m3dbnode":   buildM3DBNodeConfig,
		"dtest":      buildDTestConfig,
	}

	if _, err := os.Stat("out"); os.IsNotExist(err) {
		os.Mkdir("out", 0700)
	}
	if _, err := os.Stat("out/config"); os.IsNotExist(err) {
		os.Mkdir("out/config", 0700)
	}
	args := os.Args
	for _, arg := range args {
		// TODO: this should be passed in as a param instead
		insertFile := fmt.Sprintf("genconfig/inserts/%s-insert.yaml", arg)
		templateFile := fmt.Sprintf("genconfig/templates/%s-config.tmpl", arg)
		if buildFunc, exists := inserts[arg]; exists {
			fmt.Println("Building config for", arg)
			contents, err := buildFunc(insertFile, templateFile)
			if err == nil {
				fmt.Println(contents)
			} else {
				fmt.Println("ERROR", err)
			}
		}
	}
}

func buildM3EmAgentConfig(insertFile string, templateFile string) (string, error) {
	// Parse the insert file to an object
	insert := &inserts.M3emAgentInsert{}
	data, err := ioutil.ReadFile(insertFile)
	if err != nil {
		return "", err
	}
	if err := yaml.Unmarshal(data, insert); err != nil {
		return "", err
	}

	// open output file
	f, err := os.Create("out/config/m3em_agent.yaml")
	if err != nil {
		return "", err
	}

	// Parse the template file to receive the insert
	t := template.Must(template.ParseFiles(templateFile))
	err = t.Execute(f, insert)
	if err != nil {
		return "", err
	}
	return "Success", nil
}

func buildM3DBNodeConfig(insertFile string, templateFile string) (string, error) {
	return insertFile, errors.New("Not implemented yet for " + templateFile)
}

func buildDTestConfig(insertFile string, templateFile string) (string, error) {
	return insertFile, errors.New("Not implemented yet for " + templateFile)
}
