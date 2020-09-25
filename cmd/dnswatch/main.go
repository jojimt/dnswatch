package main

import (
	"github.com/jojimt/dnswatch/pkg/processor"
)

func main() {
	p, err := processor.NewProcessor()
	if err != nil {
		panic(err)
	}

	p.Run()
}
