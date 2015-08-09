package main

import (
	"fmt"
	"os"

	"github.com/hyangah/mgodoc/godoc"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintf(os.Stderr, "usage: %s <godoc url>\n", os.Args[0])
		os.Exit(2)
	}

	res, err := godoc.Serve(os.Args[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v", err)
		os.Exit(2)
	}

	fmt.Printf("%s", res.Body)
}
