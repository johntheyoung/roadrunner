package cmd

import (
	"fmt"
)

type VersionCmd struct{}

func (c *VersionCmd) Run(flags *RootFlags) error {
	fmt.Printf("rr version %s\n", Version)
	return nil
}
