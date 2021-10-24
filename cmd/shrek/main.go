package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"runtime"
	"strings"

	"github.com/innix/shrek"
)

type appOptions struct {
	Verbose       bool
	SaveDirectory string
	NumAddresses  int
	NumThreads    int
	Patterns      []string
}

func main() {
	opts := buildAppOptions()
	runtime.GOMAXPROCS(opts.NumThreads + 1) // +1 for main proc.

	if opts.Verbose {
		LogVerboseEnabled = true
	}

	m, err := buildMatcher(opts.Patterns)
	if err != nil {
		LogError("ERROR: invalid args: %v", err)
		os.Exit(2)
	}

	LogInfo("Searching for %d addresses, using %d threads, with %d filters.",
		opts.NumAddresses, opts.NumThreads, len(m.Inner))
	LogInfo("")

	// Channel to receive onion addresses from miners.
	addrs := make(chan *shrek.OnionAddress, opts.NumAddresses)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Spin up the miners.
	for i := opts.NumThreads; i > 0; i-- {
		go func() {
			if err := mineHostNames(ctx, addrs, m); err != nil && !errors.Is(err, ctx.Err()) {
				LogError("ERROR: %v", err)
			}
		}()
	}

	// Loop until the requested number of addresses have been mined.
	mineForever := opts.NumAddresses == 0
	for i := 0; i < opts.NumAddresses || mineForever; i++ {
		addr := <-addrs
		hostname := addr.HostNameString()

		LogInfo(hostname)
		if err := shrek.SaveOnionAddress(opts.SaveDirectory, addr); err != nil {
			LogError("ERROR: Found .onion but could not save it to file system: %v", err)
		}
	}
}

func buildAppOptions() appOptions {
	var opts appOptions
	flag.BoolVar(&opts.Verbose, "V", false, "Verbose logging (default = false)")
	flag.StringVar(&opts.SaveDirectory, "d", "", "The directory to save keys in (default = current directory)")
	flag.IntVar(&opts.NumAddresses, "n", 0, "Number of onion addresses to generate (0/default = unlimited)")
	flag.IntVar(&opts.NumThreads, "t", 0, "Number of threads to use (0/default = all CPU cores)")
	flag.Parse()

	if flag.NArg() < 1 {
		LogError("Usage: %s [COMMAND OPTIONS] <pattern1> [pattern2...]", os.Args[0])
		os.Exit(2)
	}

	// Set runtime to use number of threads requested.
	if opts.NumThreads <= 0 {
		opts.NumThreads = runtime.NumCPU()
	}

	// Set to default if negative number given for some reason.
	if opts.NumAddresses < 0 {
		opts.NumAddresses = 0
	}

	// Non-flag args are patterns.
	opts.Patterns = flag.Args()

	return opts
}

func buildMatcher(args []string) (shrek.MultiMatcher, error) {
	var mm shrek.MultiMatcher

	for _, pattern := range args {
		parts := strings.Split(pattern, ":")

		switch len(parts) {
		case 1:
			if !isValidMatcherPattern(parts[0]) {
				return mm, fmt.Errorf("pattern contains invalid chars: %q", parts[0])
			}

			mm.Inner = append(mm.Inner, shrek.StartEndMatcher{
				Start: []byte(parts[0]),
				End:   nil,
			})

			LogVerbose("Found filter: starts_with='%s'", parts[0])
		case 2:
			if !isValidMatcherPattern(parts[0]) {
				return mm, fmt.Errorf("pattern contains invalid chars: %q", parts[0])
			}
			if !isValidMatcherPattern(parts[1]) {
				return mm, fmt.Errorf("pattern contains invalid chars: %q", parts[1])
			}

			mm.Inner = append(mm.Inner, shrek.StartEndMatcher{
				Start: []byte(parts[0]),
				End:   []byte(parts[1]),
			})

			LogVerbose("Found filter: starts_with='%s', ends_with='%s'", parts[0], parts[1])
		default:
			return mm, fmt.Errorf("invalid pattern: %q", pattern)
		}
	}

	return mm, nil
}

func isValidMatcherPattern(v string) bool {
	return strings.Trim(v, "abcdefghijklmnopqrstuvwxyz234567") == ""
}

func mineHostNames(ctx context.Context, ch chan<- *shrek.OnionAddress, m shrek.Matcher) error {
	for ctx.Err() == nil {
		addr, err := shrek.MineOnionHostName(ctx, nil, m)
		if err != nil {
			return err
		}

		select {
		case ch <- addr:
		default:
			return nil
		}
	}

	return ctx.Err()
}
