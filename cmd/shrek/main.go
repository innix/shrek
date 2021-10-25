package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"runtime"
	"strings"
	"sync"

	"github.com/innix/shrek"
	"github.com/spf13/pflag"
)

const (
	appName    = "shrek"
	appVersion = "0.6.0-beta.1"
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
	wg := runWorkGroup(opts.NumThreads, func(n int) {
		if err := mineHostNames(ctx, addrs, m); err != nil && !errors.Is(err, ctx.Err()) {
			LogError("ERROR: %v", err)
		}
	})

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

	cancel()
	wg.Wait()
}

func buildAppOptions() appOptions {
	var opts appOptions
	pflag.IntVarP(&opts.NumAddresses, "onions", "n", 0, "`num`ber of onion addresses to generate (default = unlimited)")
	pflag.StringVarP(&opts.SaveDirectory, "save-directory", "d", "", "`dir`ectory to save keys in (default = current working directory)")
	pflag.IntVarP(&opts.NumThreads, "threads", "t", 0, "`num`ber of threads to use (default = all CPU cores)")
	pflag.BoolVarP(&opts.Verbose, "verbose", "V", false, "enable verbose logging")

	var help, version bool
	pflag.BoolVarP(&help, "help", "h", false, "show this help menu")
	pflag.BoolVarP(&version, "version", "v", false, "show app version")

	pflag.CommandLine.SortFlags = false
	pflag.Usage = func() {
		LogError("Usage:")
		LogError("  %s [options] filter [more-filters...]", os.Args[0])
		LogError("")
		LogError("OPTIONS")
		pflag.PrintDefaults()
	}
	pflag.Parse()

	if version {
		LogInfo("%s %s", appName, appVersion)
		os.Exit(0)
	} else if help {
		pflag.Usage()
		os.Exit(0)
	} else if pflag.NArg() < 1 {
		LogError("No filters provided.")
		LogError("")
		pflag.Usage()
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
	opts.Patterns = pflag.Args()

	return opts
}

func buildMatcher(args []string) (shrek.MultiMatcher, error) {
	var mm shrek.MultiMatcher

	for _, pattern := range args {
		parts := strings.Split(pattern, ":")

		switch len(parts) {
		case 1:
			start := parts[0]
			if !isValidMatcherPattern(start) {
				return mm, fmt.Errorf("pattern contains invalid chars: %q", start)
			}

			mm.Inner = append(mm.Inner, shrek.StartEndMatcher{
				Start: []byte(start),
				End:   nil,
			})

			LogVerbose("Found filter: starts_with='%s'", start)
		case 2:
			start, end := parts[0], parts[1]
			if !isValidMatcherPattern(start) {
				return mm, fmt.Errorf("pattern contains invalid chars: %q", start)
			}
			if !isValidMatcherPattern(end) {
				return mm, fmt.Errorf("pattern contains invalid chars: %q", end)
			}

			mm.Inner = append(mm.Inner, shrek.StartEndMatcher{
				Start: []byte(start),
				End:   []byte(end),
			})

			LogVerbose("Found filter: starts_with='%s', ends_with='%s'", start, end)
		default:
			return mm, fmt.Errorf("invalid pattern: %q", pattern)
		}
	}

	return mm, nil
}

func isValidMatcherPattern(v string) bool {
	return strings.Trim(v, "abcdefghijklmnopqrstuvwxyz234567") == ""
}

func runWorkGroup(n int, fn func(n int)) *sync.WaitGroup {
	var wg sync.WaitGroup
	wg.Add(n)

	for i := 0; i < n; i++ {
		go func(i int) {
			defer wg.Done()
			fn(i)
		}(i)
	}

	return &wg
}

func mineHostNames(ctx context.Context, ch chan<- *shrek.OnionAddress, m shrek.Matcher) error {
	for ctx.Err() == nil {
		addr, err := shrek.MineOnionHostName(ctx, nil, m)
		if err != nil {
			return err
		}

		select {
		case ch <- addr:
		case <-ctx.Done():
		}
	}

	return ctx.Err()
}
