package main

import (
	"bytes"
	"context"
	"crypto"
	_ "embed"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/cretz/bine/tor"
	"github.com/cretz/bine/torutil/ed25519"
	"github.com/innix/shrek"
)

//go:embed landing-page.html
var landingPageFileBytes []byte

func main() {
	const defaultSearchTerm = "ogre"
	const addrSearchTimeout = time.Minute * 5
	const torStartupTimeout = time.Minute * 3

	searchTerm := defaultSearchTerm
	if len(os.Args) == 2 {
		searchTerm = os.Args[1]
	}

	ctx, cancel := context.WithTimeout(context.Background(), addrSearchTimeout)
	defer cancel()

	fmt.Printf("Searching for an .onion address starting with %q.\n", searchTerm)
	fmt.Println("This could take a minute or two.")
	if len(searchTerm) > 5 {
		fmt.Println("Warning: The search term is a bit long. It could take a while to find.")
	}

	matcher := shrek.StartEndMatcher{
		Start: []byte(searchTerm),
	}

	addr, err := shrek.MineOnionHostName(ctx, nil, matcher)
	if err != nil && errors.Is(err, ctx.Err()) {
		fmt.Println("Your computer took too long trying to generate an address.")
		fmt.Println("Choose a shorter search term next time!")
		os.Exit(1)
	} else if err != nil {
		fmt.Println("Error: Could not generate an address:", err)
		os.Exit(2)
	}

	fmt.Println("Address found:", addr.HostNameString())
	fmt.Println()
	fmt.Println("Starting Tor server using address. This could take a minute.")

	t, err := tor.Start(context.Background(), nil)
	if err != nil {
		fmt.Println("Error: Could not start Tor instance:", err)
		os.Exit(3)
	}
	defer t.Close()

	// Convert the Shrek OnionAddress into a KeyPair that Bine can work with.
	kp := ed25519.PrivateKey(addr.SecretKey).KeyPair()

	onion, err := createTorListener(torStartupTimeout, t, kp)
	if err != nil {
		fmt.Println("Error: Could not create Tor listener:", err)
		os.Exit(4)
	}
	defer onion.Close()

	fmt.Printf("Open a Tor-enabled browser and browse to: http://%s.onion/\n", onion.ID)
	fmt.Println()
	fmt.Println("Press Ctrl-C to close server.")
	fmt.Println()

	if err := serveLandingPage(onion); err != nil && !errors.Is(err, http.ErrServerClosed) {
		fmt.Println("Error: Could not serve HTTP requests over Tor:", err)
		os.Exit(5)
	}
}

func createTorListener(timeout time.Duration, t *tor.Tor, key crypto.PrivateKey) (*tor.OnionService, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	onion, err := t.Listen(ctx, &tor.ListenConf{
		LocalPort:   8080,
		RemotePorts: []int{80},
		Key:         key,
	})
	if err != nil {
		return nil, err
	}

	return onion, nil
}

func serveLandingPage(l net.Listener) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		br := bytes.NewReader(landingPageFileBytes)
		http.ServeContent(w, r, "landing-page.html", time.Time{}, br)
	})

	return http.Serve(l, mux)
}
