package gogo

import (
	`fmt`
	`log`
	`os`
)

var PrintError = func(f string, args... interface{}) {
	fmt.Fprintf(os.Stderr, f, args...)
	log.Printf(f, args...)	
}