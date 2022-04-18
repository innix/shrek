package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/briandowns/spinner"
	"github.com/fatih/color"
	"github.com/innix/shrek"
	"github.com/spf13/pflag"
)

const (
	appName    = "shrek"
	appVersion = "0.6.1"
)

func main() {
	opts := buildAppOptions()
	runtime.GOMAXPROCS(opts.NumThreads + 1) // +1 for main proc.

	LogVerboseEnabled = true
	LogPrettyEnabled = opts.Formatting.UseEnhanced()
	color.NoColor = !opts.Formatting.UseColors()

	LogInfo("%sSaving found addresses to %s",
		Pretty("📁 ", ""),
		color.YellowString("%s", opts.SaveDirectory),
	)
	LogInfo("")

	m, err := buildMatcher(opts.Patterns)
	if err != nil {
		LogError("%s: Could not build search filters: %v.", color.RedString("Error"), err)
		os.Exit(2)
	}

	addrText := color.GreenString("%d", opts.NumAddresses)
	if opts.NumAddresses == 0 {
		addrText = color.GreenString("infinite")
	}
	LogInfo("%sSearching for %s addresses, using %s threads, with %s search filters:",
		Pretty("🔥 ", ""),
		addrText,
		color.GreenString("%d", opts.NumThreads),
		color.GreenString("%d", len(m.Inner)),
	)
	defer func() {
		LogInfo("")
		LogInfo("%sShrek has finished searching.", Pretty("👍 ", ""))
	}()

	// Channel to receive onion addresses from miners.
	addrs := make(chan *shrek.OnionAddress, opts.NumAddresses)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Spin up the miners.
	wg := runWorkGroup(opts.NumThreads, func(n int) {
		if err := mineHostNames(ctx, addrs, m); err != nil && !errors.Is(err, ctx.Err()) {
			LogError("%s: %v.", color.RedString("Error"), err)
		}
	})

	// Loop until the requested number of addresses have been mined.
	mineForever := opts.NumAddresses == 0
	ps := newProgressSpinner("   ", time.Millisecond*130)
	for i := 0; i < opts.NumAddresses || mineForever; i++ {
		ps.Start()
		addr := <-addrs
		hostname := addr.HostNameString()
		ps.Stop()

		LogInfo("%s%s", Pretty("   🔹 ", ""), hostname)
		if err := shrek.SaveOnionAddress(opts.SaveDirectory, addr); err != nil {
			LogError("%s: Found .onion but could not save it to file system: %v.",
				color.RedString("Error"),
				err,
			)
		}
	}

	cancel()
	wg.Wait()
}

func buildAppOptions() appOptions {
	var opts appOptions

	pflag.IntVarP(&opts.NumAddresses, "onions", "n", 0, "`num`ber of onion addresses to generate, 0 = infinite (default = 1)")
	pflag.StringVarP(&opts.SaveDirectory, "save-dir", "d", "", "`dir`ectory to save addresses in (default = cwd)")
	pflag.IntVarP(&opts.NumThreads, "threads", "t", 0, "`num`ber of threads to use (default = all CPU cores)")
	pflag.VarP(&opts.Formatting, "format", "", "what `kind` of formatting to use (basic, colored, enhanced, default = all)")

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

	// Set non-zero defaults here to prevent pflag from printing the default values itself.
	// It can't be disabled and we want to print it differently from how pflag does it.
	if f := pflag.Lookup("onions"); !f.Changed {
		if err := f.Value.Set("1"); err != nil {
			panic(err)
		}
	}

	if version {
		LogInfo("%s %s, os: %s, arch: %s", appName, appVersion, runtime.GOOS, runtime.GOARCH)
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
			LogError("%s: Could not resolve save dir to absolute path: %v.",
				color.RedString("Error"),
				err,
			)
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
	var inner []shrek.StartEndMatcher

	for _, pattern := range args {
		parts := strings.Split(pattern, ":")

		switch len(parts) {
		case 1:
			start := parts[0]
			m := shrek.StartEndMatcher{
				Start: []byte(start),
				End:   nil,
			}
			if err := m.Validate(); err != nil {
				return mm, fmt.Errorf(
					"pattern '%s' is not valid: %w", color.YellowString("%s", pattern), err,
				)
			}
			inner = append(inner, m)
		case 2:
			start, end := parts[0], parts[1]
			m := shrek.StartEndMatcher{
				Start: []byte(start),
				End:   []byte(end),
			}
			if err := m.Validate(); err != nil {
				return mm, fmt.Errorf(
					"pattern '%s' is not valid: %w", color.YellowString("%s", pattern), err,
				)
			}
			inner = append(inner, m)
		default:
			return mm, fmt.Errorf(
				"pattern '%s' is not a valid syntax", color.YellowString("%s", pattern),
			)
		}
	}

	LogVerbose("%sLooking for addresses that match any of these conditions:", Pretty("🔎 ", ""))
	for _, m := range inner {
		startsWith := fmt.Sprintf("'%s'", color.YellowString("%s", m.Start))
		endsWith := fmt.Sprintf("'%s'", color.YellowString("%s", m.End))
		if len(m.Start) == 0 {
			startsWith = color.YellowString("anything")
		}
		if len(m.End) == 0 {
			endsWith = color.YellowString("anything")
		}

		LogVerbose("%sAn address that starts with %s and ends with %s",
			Pretty("   🔸 ", " - "),
			startsWith,
			endsWith,
		)

		mm.Inner = append(mm.Inner, m)
	}
	LogVerbose("")

	return mm, nil
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

func newProgressSpinner(prefix string, speed time.Duration) *spinner.Spinner {
	s := spinner.New(spinner.CharSets[14], speed)
	s.HideCursor = true
	s.Prefix = prefix
	s.Suffix = " "
	s.Writer = os.Stderr

	if !LogPrettyEnabled {
		s.Delay = time.Second * 30
		s.HideCursor = false
		s.Writer = io.Discard
	}

	return s
}
