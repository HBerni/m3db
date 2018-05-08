package genconfig

import (
	"io/ioutil"

	yaml "gopkg.in/yaml.v2"
)

type M3emAgentInsert struct {
	Port       string `yaml:"port"`
	DebugPort  string `yaml:"debug_port"`
	M3Address  string `yaml:"m3_address"`
	WorkingDir string `yaml:"working_dir"`
}

func (i *M3emAgentInsert) ReadFromFile(insertFile string) error {
	data, err := ioutil.ReadFile(insertFile)
	if err != nil {
		return err
	}
	if err := yaml.Unmarshal(data, i); err != nil {
		return err
	}
	return nil
}
