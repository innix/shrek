package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"github.com/fatih/color"
	"github.com/innix/shrek"
	"github.com/spf13/pflag"
)

const (
	appName    = "shrek"
	appVersion = "0.6.0-beta.2"
)

type appOptions struct {
	NumAddresses  int
	SaveDirectory string
	NumThreads    int
	NoColors      bool
	Verbose       bool
	Patterns      []string
}

func main() {
	opts := buildAppOptions()
	runtime.GOMAXPROCS(opts.NumThreads + 1) // +1 for main proc.

	if opts.Verbose {
		LogVerboseEnabled = true
	}
	color.NoColor = opts.NoColors

	m, err := buildMatcher(opts.Patterns)
	if err != nil {
		LogError("ERROR: invalid args: %v", err)
		os.Exit(2)
	}

	addrText := color.GreenString("%d", opts.NumAddresses)
	if opts.NumAddresses == 0 {
		addrText = color.GreenString("infinite")
	}
	LogInfo("Searching for %s addresses, using %s threads, with %s filters.",
		addrText,
		color.GreenString("%d", opts.NumThreads),
		color.GreenString("%d", len(m.Inner)),
	)
	LogInfo("Saving found addresses to: '%s'", color.YellowString("%s", opts.SaveDirectory))
	LogInfo("")
	defer func() {
		LogInfo("")
		LogInfo("Shrek has finished mining.")
	}()

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
			LogError("ERROR: found .onion but could not save it to file system: %v", err)
		}
	}

	cancel()
	wg.Wait()
}

func buildAppOptions() appOptions {
	var opts appOptions
	pflag.IntVarP(&opts.NumAddresses, "onions", "n", 0, "`num`ber of onion addresses to generate (default = infinite)")
	pflag.StringVarP(&opts.SaveDirectory, "save-dir", "d", "", "`dir`ectory to save addresses in (default = cwd)")
	pflag.IntVarP(&opts.NumThreads, "threads", "t", 0, "`num`ber of threads to use (default = all CPU cores)")
	pflag.BoolVarP(&opts.NoColors, "no-colors", "", false, "disable colored console output")
	pflag.BoolVarP(&opts.Verbose, "verbose", "V", false, "enable verbose logging")

	var help, version bool
	pflag.BoolVarP(&help, "help", "h", false, "show this help menu")
	pflag.BoolVarP(&version, "version", "v", false, "show app version")

	pflag.CommandLine.SortFlags = false
	pflag.Usage = func() {
		LogError("Usage:")
		LogError("  %s [options] filter [more-filters...]", filepath.Base(os.Args[0]))
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

	// Translate save dir to absolute dir. Serves no logical purpose, just improves logging.
	if !filepath.IsAbs(opts.SaveDirectory) {
		absd, err := filepath.Abs(opts.SaveDirectory)
		if err != nil {
			LogError("ERROR: could not resolve save dir to absolute path: %v", err)
			os.Exit(1)
		}
		opts.SaveDirectory = absd
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

			LogVerbose("Found filter: starts_with='%s'", color.YellowString("%s", start))
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

			LogVerbose("Found filter: starts_with='%s', ends_with='%s'",
				color.YellowString("%s", start),
				color.YellowString("%s", end),
			)
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
