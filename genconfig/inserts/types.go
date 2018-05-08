package genconfig

type M3emAgentInsert struct {
	Port       string `yaml:"port"`
	DebugPort  string `yaml:"debug_port"`
	M3Address  string `yaml:"m3_address"`
	WorkingDir string `yaml:"working_dir"`
}
