package genconfig

type Insert interface {
	ReadFromFile(insertFile string) error
}
