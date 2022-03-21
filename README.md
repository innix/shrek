<div align="center">

# Shrek <img src="./assets/onion-icon.png" alt=":onion:" title=":onion:" class="emoji" height="20">

[![Go Reference   ][goref-badge]][goref-page]&nbsp;
[![GitHub Release ][ghrel-badge]][ghrel-page]&nbsp;
[![Project License][licen-badge]][licen-page]&nbsp;
<!-- [![Go Report      ][gorep-badge]][gorep-page]&nbsp; -->
<!-- [![Go Version     ][gover-badge]][gover-page]&nbsp; -->

Shrek is a vanity `.onion` address generator.

Runs on Linux, macOS, Windows, and more. A single static binary with no dependencies.

![Shrek running from CLI](./assets/shrek-session.webp)

</div>

# Usage (CLI)

Shrek compiles to a single binary that can be used on the CLI. Download a pre-compiled
build from the [GitHub Releases][ghrel-page] page. Or build it from source:

```bash
go install github.com/innix/shrek/cmd/shrek@latest
```

This will place the compiled shrek binary in `$GOPATH/bin`.

The program takes 1 or more filters as arguments. It only generates v3 `.onion` addresses.

```bash
# Generate an address that starts with "food":
shrek food

# Generate an address that starts with "food" and ends with "xid":
shrek food:xid

# Generate an address that starts with "food" and ends with "xid", or starts with "barn":
shrek food:xid barn

# Generate an address that ends with "2ayd".
shrek :2ayd

# Shrek can search for the start of an onion address much faster than the end of the
# address. Therefore, it is recommended that the filters you use have a bigger start
# filter and a smaller (or zero) end filter.
```

To see full usage, use the help flag `-h`:

```bash
shrek -h
```

# Usage (library)

You can use Shrek as a library in your Go code. Add it to your `go.mod` file by running
this in your project root:

```bash
go get github.com/innix/shrek
```

Here's an example of using Shrek to find an address using a start-end pattern and
saving it to disk:

```go
package main

import (
	"context"
	"fmt"

	"github.com/innix/shrek"
)

func main() {
	// Brute-force find a hostname that starts with "foo" and ends with "id".
	addr, err := shrek.MineOnionHostName(context.Background(), nil, shrek.StartEndMatcher{
		Start: []byte("foo"),
		End:   []byte("id"),
	})
	if err != nil {
		panic(err)
	}

	// Save hostname and the public/secret keys to disk.
	fmt.Printf("Onion address %q found, saving to file system.\n", addr.HostNameString())
	err = shrek.SaveOnionAddress("output_dir", addr)
	if err != nil {
		panic(err)
	}
}
```

More comprehensive examples of how to use Shrek as a library can be found in
the [examples](./examples) directory.

# In active development

This project is under active development and hasn't reached `v1.0` yet. Therefore the public
API is not fully stable and may contain breaking changes as new versions are released. If
you're using Shrek as a package in your code and an update breaks it, feel free to open an
issue and a contributor will help you out.

Once Shrek reaches `v1.0`, the API will stabilize and any new changes will not break your
existing code.

# Performance and goals

Shrek is the fastest `.onion` vanity address generator written in Go (at time of writing), but
it's still slow compared to the mature and highly optimized [`mkp224o`][mkp224o-page] program.
There are optimizations that could be made to Shrek to improve its performance. Contributions
are welcome and encouraged. Feel free to open an issue/discussion to share your thoughts and
ideas, or submit a pull request with your optimization.

The primary goal of Shrek is to be an easy to use Go program/library for Go developers,
not to be the fastest program out there. It should be able to run on every platform that
Go supports. Use of `cgo` or other complicated build processes should be avoided.

# FAQ

## Are v2 addresses supported?

No. They were already deprecated at the time Shrek was written, so there didn't seem any
point supporting them. There are no plans to add it as a feature.

## Can I run Shrek without all the emojis, extra colors, and other fancy formatting?

Sure. Use the `--format basic` flag when running the program to keep things simple. You
can also run `shrek --help` to see a list of all possible formatting options; maybe you'll
find one you like.

## How do I use a generated address with the `cretz/bine` Tor library?

There are [example projects](./examples) that show how to use Shrek and Bine together.

But to put it simply, all you need to do is convert Shrek's `OnionAddress` into Bine's
`KeyPair`. See this code example:

<details>
  <summary>Click to expand answer</summary>

```go
package main

import (
    "github.com/cretz/bine/tor"
    "github.com/cretz/bine/torutil/ed25519"
    "github.com/innix/shrek"
)

func main() {
    // Generate any address, the value doesn't matter for this demo.
    addr, err := shrek.GenerateOnionAddress(nil)
    if err != nil {
        panic(err)
    }

    // Or read a previously generated address that was saved to disk with SaveOnionAddress.
    // addr, err := shrek.ReadOnionAddress("./addrs/bqyql3bq532kzihcmp3c6lb6id.onion/")
    // if err != nil {
    //     panic(err)
    // }

    // Take the private key from Shrek's OnionAddress and turn it into an ed25519.KeyPair
    // that the Bine library can understand.
    keyPair := ed25519.PrivateKey(addr.SecretKey).KeyPair()

    // Now you can use the KeyPair in Bine as you normally would, e.g. with ListenConf:
    listenConf := &tor.ListenConf{
        Key: keyPair,
    }
}
```
</details>

## Why "Shrek"?

Onions have layers, ogres have layers.

# License

Shrek is distributed under the terms of the MIT License (see [LICENSE](LICENSE)).


<!-- Link refs -->
[goref-badge]: <https://img.shields.io/badge/-reference-007d9c?style=for-the-badge&logo=go&labelColor=5c5c5c&logoColor=ffffff> "Go pkg.dev badge"
[goref-page]: <https://pkg.go.dev/github.com/innix/shrek> "Go pkg.dev"

[ghrel-badge]: <https://img.shields.io/github/v/release/innix/shrek?display_name=tag&sort=semver&style=for-the-badge> "GitHub Releases badge"
[ghrel-page]: <https://github.com/innix/shrek/releases> "GitHub Releases"

[gorep-badge]: <https://goreportcard.com/badge/github.com/innix/shrek?style=for-the-badge&logo=go> "Go Report badge"
[gorep-page]: <https://goreportcard.com/report/github.com/innix/shrek> "Go Report"

[gover-badge]: <https://img.shields.io/github/go-mod/go-version/innix/shrek?style=for-the-badge&logo=go> "Go Version badge"
[gover-page]: <go.mod> "Go Version"

[licen-badge]: <https://img.shields.io/github/license/innix/shrek?style=for-the-badge> "Project License badge"
[licen-page]: <LICENSE> "Project License"

[mkp224o-page]: <https://github.com/cathugger/mkp224o> "cathugger/mkp224o - GitHub page"
