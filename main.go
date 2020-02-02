package main

import (
	"fmt"
	"io"
	"strings"
)

func main() {
	/* Token Client */
	printSeprator := func(i int) {
		fmt.Printf("-------------------------------------- %d --------------------------------------\n", i)
	}

	tokenizer := NewTokenizer(strings.NewReader(inpContent))
	var err error
	for i := 0; err != io.EOF; i++ {
		tokenizer.Next()
		token := tokenizer.Token()
		if token.Type == ErrorToken {
			err = tokenizer.Err()
			printSeprator(i)
			fmt.Printf("%#v. err=%v\n", token, err)
			continue
		}
		if s := token.String(); len(s) > 0 {
			s = strings.ReplaceAll(s, "\n", "↵\n")
			s = strings.ReplaceAll(s, " ", "␣")
			printSeprator(i)
			fmt.Printf("%v\n", token)
		} else {
			printSeprator(i)
			fmt.Printf("%#v\n", token)
		}
	}
}

func preprocess() {
	// rm "**" lines
	// output "pre_processed.inp"
}
