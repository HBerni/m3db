package genconfig

import (
	"errors"
)

type M3DBNodeInsert struct{}

func (i *M3DBNodeInsert) ReadFromFile(insertFile string) error {
	return errors.New("Not yet implemented")
}
