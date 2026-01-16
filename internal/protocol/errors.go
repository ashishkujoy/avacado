package protocol

import "fmt"

func NewUnknowCommandError(name string) error {
	return fmt.Errorf("unknown command %s", name)
}
