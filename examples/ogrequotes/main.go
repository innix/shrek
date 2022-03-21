package main

import (
	"bytes"
	"context"
	"crypto"
	"errors"
	"fmt"
	"math/rand"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/cretz/bine/tor"
	"github.com/cretz/bine/torutil/ed25519"
	"github.com/innix/shrek"
)

func main() {
	const searchTerm = "ogre"
	const addrSearchTimeout = time.Minute * 5
	const torStartupTimeout = time.Minute * 3

	ctx, cancel := context.WithTimeout(context.Background(), addrSearchTimeout)
	defer cancel()

	fmt.Printf("Searching for an .onion address starting with %q.\n", searchTerm)
	fmt.Println("This could take a minute or two.")

	matcher := shrek.StartEndMatcher{
		Start: []byte(searchTerm),
	}

	addr, err := shrek.MineOnionHostName(ctx, nil, matcher)
	if err != nil && errors.Is(err, ctx.Err()) {
		fmt.Println("Your computer took too long trying to generate an address.")
		fmt.Println("Choose a shorter search term next time!")
		os.Exit(1)
	} else if err != nil {
		fmt.Println("Error: Could not generate the .onion address:", err)
		os.Exit(2)
	}

	fmt.Println("The .onion address found:", addr.HostNameString())
	fmt.Println()
	fmt.Println("Starting Tor server using found address. This could take a minute.")

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

	if err := serveOgreQuotes(onion); err != nil && !errors.Is(err, http.ErrServerClosed) {
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

func serveOgreQuotes(l net.Listener) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/", getQuoteHandler())

	return http.Serve(l, mux)
}

func getQuoteHandler() http.HandlerFunc {
	ogreQuotes, err := loadOgreQuotes()
	if err != nil {
		panic(fmt.Errorf("could not load ogre quotes: %w", err))
	}

	return func(w http.ResponseWriter, r *http.Request) {
		// Pick random quote.
		n := rand.Intn(len(ogreQuotes))
		lines := ogreQuotes[n]

		// Create buffer and write HTML header to it.
		bb := &bytes.Buffer{}
		htmlHeader := fmt.Sprintf("<h1>Ogre Quote #%d</h1><br/>\n\n", n+1)
		bb.WriteString(htmlHeader)

		// Write quote in HTML form to the buffer.
		for _, l := range lines {
			if len(l.Speaker) > 0 {
				bb.WriteString(fmt.Sprintf("<b>%s:</b> ", l.Speaker))
			}
			bb.WriteString(fmt.Sprintf("%s<br/>\n", l.Line))
		}

		// Serve the buffer content to client.
		fmt.Printf("Serving ogre quote #%d to %s\n", n+1, r.UserAgent())
		http.ServeContent(w, r, "index.html", time.Time{}, bytes.NewReader(bb.Bytes()))
	}
}
