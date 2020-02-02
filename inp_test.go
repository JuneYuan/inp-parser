package main

import (
	"fmt"
	"io"
	"strings"
	"testing"

	"golang.org/x/net/html"
)

var printSeprator = func(i int) {
	fmt.Printf("-------------------------------------- %d --------------------------------------\n", i)
}

// TestInpTokenizer is a Client of Tokenizer
// output: inp-tokens.txt
func TestInpTokenizer(t *testing.T) {
	tokenizer := NewTokenizer(strings.NewReader(inpContent))
	var err error
	for i := 0; err != io.EOF; i++ {
		tokenizer.Next()
		token := tokenizer.Token()
		if token.Type == ErrorToken {
			err = tokenizer.Err()
			printSeprator(i)
			fmt.Printf("%#v. err=%v \n", token, err)
			continue
		}
		if s := token.String(); len(s) > 0 {
			s = strings.ReplaceAll(s, "\n", "↵\n")
			s = strings.ReplaceAll(s, " ", "␣")
			printSeprator(i)
			fmt.Printf("%v", s)
		} else {
			printSeprator(i)
			fmt.Printf("%#v", token)
		}
	}
}

// TestHTMLTokenizer ...
// output: html-tokens.txt
func TestHTMLTokenizer(t *testing.T) {
	// if .String() available, print it
	// replace "\n" with "↵\n"
	// replace " " with "␣"
	// if .String() not available, print "#v"
	tokenizer := html.NewTokenizer(strings.NewReader(basicHTMLDoc))
	var err error
	for i := 0; err != io.EOF; i++ {
		tokenizer.Next()
		token := tokenizer.Token()
		if token.Type == html.ErrorToken {
			err = tokenizer.Err()
			printSeprator(i)
			fmt.Printf("%#v. err=%v\n", token, err)
			continue
		}
		if s := token.String(); len(s) > 0 {
			s = strings.ReplaceAll(s, "\n", "↵\n")
			s = strings.ReplaceAll(s, " ", "␣")
			printSeprator(i)
			fmt.Printf("%v", s)
			if !strings.HasSuffix(s, "\n") {
				fmt.Println()
			}
		} else {
			printSeprator(i)
			fmt.Printf("%#v\n", token)
		}
	}
}
