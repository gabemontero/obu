package util

import (
	"fmt"
	"os"
)

func DefaultMessage() {
	fmt.Fprintf(os.Stdout, "Use one of the available options listed under help to get\n")
	fmt.Fprintf(os.Stdout, "content that can be stored in files, parameters, or environment\n")
	fmt.Fprintf(os.Stdout, "variables that can be subsequently consumed by your image\n")
	fmt.Fprintf(os.Stdout, "build tool.\n")
}
