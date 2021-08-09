// Copyright (c) 2021 Fract
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package main

import (
	"bufio"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/fract-lang/fract/parser"
	"github.com/fract-lang/fract/pkg/fract"
	"github.com/fract-lang/fract/pkg/obj"
	"github.com/fract-lang/fract/pkg/str"
)

func getNamespace(cmd string) string {
	i := strings.IndexByte(cmd, ' ')
	if i == -1 {
		return cmd
	}
	return cmd[0:i]
}

func removeNamespace(cmd string) string {
	i := strings.IndexByte(cmd, ' ')
	if i == -1 {
		return ""
	}
	return cmd[i+1:]
}

func input(msg string) string {
	fmt.Print(msg)
	//! Don't use fmt.Scanln
	s := bufio.NewScanner(os.Stdin)
	s.Scan()
	return s.Text()
}

var p *parser.Parser

func interpret() {
	for {
		p.Lex.File.Lines = []string{input(">> ")}
	reTokenize:
		p.Tokens = nil
	reTokenizeUnNil:
		p.Lex.Finished = false
		p.Lex.Braces = 0
		p.Lex.Brackets = 0
		p.Lex.Parentheses = 0
		p.Lex.RangeComment = false
		p.Lex.Line = 1
		p.Lex.Column = 1
		/* Tokenize all lines. */
		for !p.Lex.Finished {
			tks := p.Lex.Next()
			// Check multiline comment.
			if p.Lex.RangeComment {
				p.Lex.File.Lines = append(p.Lex.File.Lines, []string{input(" | ")}...)
				goto reTokenizeUnNil
			}
			// cacheTokens are empty?
			if tks == nil {
				continue
			}
			// Check parentheses.
			if p.Lex.Braces > 0 || p.Lex.Brackets > 0 || p.Lex.Parentheses > 0 {
				p.Lex.File.Lines = append(p.Lex.File.Lines, []string{input(" | ")}...)
				goto reTokenize
			}
			p.Tokens = append(p.Tokens, tks)
		}
		p.Interpret()
	}
}

func catch(e obj.Panic) {
	if e.Msg == "" {
		return
	}
	fmt.Println(e.Msg)
}

func help(cmd string) {
	if cmd != "" {
		fmt.Println("This module can only be used!")
		return
	}
	helpMap := map[string]string{
		"version": "Show version.",
		"help":    "Show help.",
	}
	maxKeyLen := 0
	for k := range helpMap {
		if maxKeyLen < len(k) {
			maxKeyLen = len(k)
		}
	}
	maxKeyLen += 5
	for k := range helpMap {
		fmt.Println(k + " " + str.Full(maxKeyLen-len(k), ' ') + helpMap[k])
	}
}

func version(cmd string) {
	if cmd != "" {
		fmt.Println("This module can only be used!")
		return
	}
	fmt.Println("Fract Version [" + fract.Version + "]")
}

// make module is interpret source file.
func make(cmd string) {
	if cmd == "" {
		fmt.Println("This module cannot only be used!")
		return
	} else if !strings.HasSuffix(cmd, fract.Extension) {
		cmd += fract.Extension
	}
	if info, err := os.Stat(cmd); err != nil || info.IsDir() {
		fmt.Println("The Fract file is not exists: " + cmd)
		return
	}
	p := parser.New(cmd)
	p.AddBuiltInFuncs()
	(&obj.Block{
		Try: p.Interpret,
		Catch: func(e obj.Panic) {
			os.Exit(0)
		},
	}).Do()
}

// makeCheck is check command is valid source code path or not.
func makeCheck(path string) bool {
	if strings.HasSuffix(path, fract.Extension) {
		return true
	}
	info, err := os.Stat(path + fract.Extension)
	return err == nil && !info.IsDir()
}

func processCommand(namespace, cmd string) {
	switch namespace {
	case "help":
		help(cmd)
	case "version":
		version(cmd)
	default:
		if makeCheck(namespace) {
			make(namespace)
		} else {
			fmt.Println("There is no such command!")
		}
	}
}

func init() {
	fract.ExecutablePath = filepath.Dir(os.Args[0])
	// Check standard library.
	if info, err := os.Stat(path.Join(fract.ExecutablePath, fract.StdLib)); err != nil || !info.IsDir() {
		fmt.Println("Standard library not found!")
		input("\nPress enter for exit...")
		os.Exit(1)
	}
	// Not started with arguments.
	if len(os.Args) < 2 {
		return
	}

	defer os.Exit(0)
	var sb strings.Builder
	for _, arg := range os.Args[1:] {
		sb.WriteString(" " + arg)
	}
	os.Args[0] = sb.String()[1:]
	processCommand(getNamespace(os.Args[0]), removeNamespace(os.Args[0]))
}

func main() {
	fmt.Println("Fract " + fract.Version + " (c) MIT License.\n" + "Fract Developer Team.\n")
	fract.InteractiveShell = true
	p = parser.NewStdin()
	p.AddBuiltInFuncs()
	b := &obj.Block{
		Try:   interpret,
		Catch: catch,
	}
	for {
		b.Do()
	}
}
