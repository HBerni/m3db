package main

import (
	"fmt"
	"os"
	"text/template"

	inserts "github.com/m3db/m3db/genconfig/inserts"
)

var insertMap = map[string]inserts.Insert{
	"m3em_agent": &inserts.M3emAgentInsert{},
	"m3dbnode":   &inserts.M3DBNodeInsert{},
	"dtest":      &inserts.DTestInsert{},
}

type BuildOptions struct {
	insertFile   string
	templateFile string
	outputFile   string
	insert       inserts.Insert
}

func NewBuildOptions(service string, i inserts.Insert) BuildOptions {
	return BuildOptions{
		insertFile:   insertFilePath(service),
		templateFile: templateFilePath(service),
		outputFile:   outputFilePath(service),
		insert:       i,
	}
}

// TODO: instead of using a yaml file, the insert can probably defined at
// a global level and have default settings per environment
func main() {

	// extract later
	if _, err := os.Stat("out"); os.IsNotExist(err) {
		os.Mkdir("out", 0700)
	}
	if _, err := os.Stat("out/config"); os.IsNotExist(err) {
		os.Mkdir("out/config", 0700)
	}

	// parse args
	args := os.Args
	for _, arg := range args {
		if insert, exists := insertMap[arg]; exists {
			fmt.Println("Building config for", arg)
			opts := NewBuildOptions(arg, insert)
			err := buildConfig(opts)
			if err == nil {
				fmt.Println("Successfully generated the config for", arg)
			} else {
				fmt.Println("ERROR", err)
			}
		}
	}
}

func insertFilePath(service string) string {
	return fmt.Sprintf("genconfig/inserts/%s-insert.yaml", service)
}

func templateFilePath(service string) string {
	return fmt.Sprintf("genconfig/templates/%s-config.tmpl", service)
}

func outputFilePath(service string) string {
	return fmt.Sprintf("out/config/%s.yaml", service)
}

func buildConfig(opts BuildOptions) error {
	// parse insert
	err := opts.insert.ReadFromFile(opts.insertFile)

	// open output file
	f, err := os.Create(opts.outputFile)
	if err != nil {
		return err
	}

	// Parse the template file to receive the insert
	t := template.Must(template.ParseFiles(opts.templateFile))
	err = t.Execute(f, opts.insert)
	if err != nil {
		return err
	}
	return nil
}
