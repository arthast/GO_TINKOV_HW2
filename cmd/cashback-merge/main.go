// Package main implements the Cashback Merge CLI application.
package main

import (
	"bytes"
	"cashbackmerge/internal/game"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
)

func main() {
	if err := run(os.Args[1:], os.Stdout, os.Stderr); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(args []string, stdout, stderr io.Writer) error {
	flags := flag.NewFlagSet("cashback-merge", flag.ContinueOnError)
	flags.SetOutput(stderr)
	configPath := flags.String("config", "", "path to campaign JSON")
	seed := flags.Uint64("seed", 341124, "random seed")
	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}
	if *configPath == "" || flags.NArg() != 0 {
		return fmt.Errorf("usage: cashback-merge --config campaign.json --seed 341124")
	}
	data, err := os.ReadFile(*configPath)
	if err != nil {
		return fmt.Errorf("read campaign: %w", err)
	}
	cfg, err := game.LoadConfig(bytes.NewReader(data))
	if err != nil {
		return err
	}
	g, err := game.New(cfg, *seed)
	if err != nil {
		return err
	}
	_, err = fmt.Fprint(stdout, g.State)
	return err
}
