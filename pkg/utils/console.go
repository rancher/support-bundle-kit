package utils

import (
	"bufio"
	"fmt"
	"os"
)

func WriteStdout(msg string) {
	stdout := bufio.NewWriter(os.Stdout)
	_, _ = fmt.Fprint(stdout, msg)
	_ = stdout.Flush()
}
