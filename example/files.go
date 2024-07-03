package main

import (
	"fmt"
	"time"

	"khan.rip"
)

func init() {
	for i := 0; i < 100; i++ {
		_ = time.Now()
		khan.Add(&khan.File{
			Path: fmt.Sprintf("/tmp/file_%d", i),
			//Content: fmt.Sprintf("Contents of file %d is neat: %s\n", i, time.Now()),
			Content: fmt.Sprintf("Contents of file %d is neat: %s\n", i, "heh"),
		})
	}
}
