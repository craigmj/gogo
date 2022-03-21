package gogo

import (
	`fmt`
	`log`
	`os`
)

func PrintError(f string, args... interface{}) {
	fmt.Fprintf(os.Stderr, f, args...)
	log.Printf(os.Stderr, f, args...)	
}