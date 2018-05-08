package genconfig

import (
	"errors"
)

type DTestInsert struct{}

func (i *DTestInsert) ReadFromFile(insertFile string) error {
	return errors.New("Not yet implemented")
}
