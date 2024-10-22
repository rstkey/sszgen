// Copyright 2023 sszgen authors
// SPDX-License-Identifier: BSD-3-Clause

package main

import (
	"bytes"
	"flag"
	"fmt"
	"go/format"
	"os"

	"golang.org/x/tools/go/packages"
)

func main() {
	var (
		pkgdir   = flag.String("dir", ".", "input package")
		output   = flag.String("out", "-", "output file (default is stdout)")
		typename = flag.String("type", "", "type to generate methods for")
	)
	flag.Parse()

	cfg := Config{
		Dir:  *pkgdir,
		Type: *typename,
	}
	code, err := cfg.process()
	if err != nil {
		fatal(err)
	}
	if *output == "-" {
		os.Stdout.Write(code)
	} else if err := os.WriteFile(*output, code, 0600); err != nil {
		fatal(err)
	}
}

func fatal(args ...interface{}) {
	fmt.Fprintln(os.Stderr, args...)
	os.Exit(1)
}

type Config struct {
	Dir  string // input package directory
	Type string
}

// process generates the Go code.
func (cfg *Config) process() ([]byte, error) {
	// Load packages.
	pcfg := &packages.Config{
		Mode:       packages.NeedName | packages.NeedTypes | packages.NeedImports | packages.NeedDeps,
		Dir:        cfg.Dir,
		BuildFlags: []string{"-tags", "nosszgen"},
	}
	ps, err := packages.Load(pcfg)
	if err != nil {
		return nil, err
	}
	if len(ps) == 0 {
		return nil, fmt.Errorf("no Go package found in %s", cfg.Dir)
	}
	if len(ps) != 1 {
		return nil, fmt.Errorf("at most one package can be processed at the same time")
	}
	packages.PrintErrors(ps)

	pkg := ps[0]
	if len(pkg.Errors) > 0 {
		return nil, fmt.Errorf("package %s has errors", pkg.PkgPath)
	}
	types, err := parsePackage(pkg.Types, nil)
	if err != nil {
		return nil, err
	}
	var (
		ctx    = newGenContext(pkg.Types)
		chunks [][]byte
	)
	for _, typ := range types {
		ret, err := generate(ctx, typ)
		if err != nil {
			return nil, err
		}
		chunks = append(chunks, ret)
	}
	code := bytes.Join(chunks, []byte("\n\n"))

	// Add package and imports definition and format code
	code = append(ctx.header(), code...)
	code, _ = format.Source(code)

	// Add build comments.
	// This is done here to avoid processing these lines with gofmt.
	var header bytes.Buffer
	fmt.Fprint(&header, "// Code generated by sszgen. DO NOT EDIT.\n\n")
	fmt.Fprint(&header, "//go:build !nosszgen\n")
	fmt.Fprint(&header, "// +build !nosszgen\n\n")
	return append(header.Bytes(), code...), nil
}
