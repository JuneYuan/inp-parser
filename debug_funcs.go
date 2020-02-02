package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

var (
	szLmt = int64(20)
)

// getShortInpContent prints the .inp content, keeping only szLmt lines of data lines
// output: short_blast_cell.inp
func getShortInpContent() {
	// scan by line
	// trim heading and tailing
	// try to split by "," and find the first token, parse as int

	if len(os.Args) > 1 {
		szLmt, _ = strconv.ParseInt(os.Args[1], 10, 64)
	}

	input := bufio.NewScanner(os.Stdin)
	for input.Scan() {
		raw := input.Text()
		line := strings.Trim(raw, " ")
		delLine := false

		if tokens := strings.Split(line, ","); len(tokens) > 0 {
			s := tokens[0]

			if strings.HasPrefix(s, "Euler") {
				s = s[8:]
			}

			if seqNo, err := strconv.ParseInt(s, 10, 64); err == nil && seqNo > szLmt {
				delLine = true
			}
		}

		if delLine {
			continue
		}
		fmt.Println(raw)
	}
}
