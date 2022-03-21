# Shrek code examples

This directory contains a few example projects that show how to use Shrek as a
library in your Go project.

# Coal Miner

The Coal Miner project is a very basic vanity `.onion` address finder. It searches
for 5 `.onion` addresses that start with the word "coal". It outputs each address
to the console once it finds it, then exits once all 5 have been found. It only uses
a single goroutine, which means there's no concurrency; it only searches for one
address at a time. This was done intentionally to keep the code simple.

To start the program, open a console in this directory and run:

```bash
go run ./coalminer/
```

# Hello World on Tor

🧅 _You'll need Tor installed on your computer to run this example._

The Hello World project is a very basic HTTP server that runs on the Tor network. It
hosts a [single static HTML file](./helloworld/landing-page.html), which contains
some very basic information about the Shrek project.

When the program is started, it will use Shrek to generate an `.onion` address. You
can specify what the address should start with by passing a string to the program as
the first argument. If you don't provide one, then "ogre" will be used by default.
The address might take a minute or two to generate, depending on how powerful your
computer is.

After an address has been found, the program will use it to configure an HTTP server
to run on the Tor network. Once connected to the Tor network, the program will output
a full `.onion` URL to the console. Copy and paste it into a Tor Browser, and you
should see the `landing-page.html` page load. That link will work for everyone who
has a Tor Browser.

To start the program, open a console in this directory and run:

```bash
# Search for .onion address that starts with default value "ogre".
go run ./helloworld/

# Search for .onion address that starts with custom value "fizz".
go run ./helloworld/ fizz
```

# Random Ogre Quotes on Tor

🧅 _You'll need Tor installed on your computer to run this example._

The Random Ogre Quotes project is another basic HTTP server that runs on the Tor
network. It hosts a single HTML page which contains a random quote from the Shrek
movie. The quote changes everytime you refresh the page.

When the program is started, it will use Shrek to generate an `.onion` address that
starts with "ogre". The address might take a minute or two to generate, depending
on how powerful your computer is.

After an address has been found, the program will use it to configure an HTTP server
to run on the Tor network. Once connected to the Tor network, the program will output
a full `.onion` URL to the console. Copy and paste it into a Tor Browser, and you
should see the page load with a random Shrek quote. That link will work for everyone
who has a Tor Browser.

To start the program, open a console in this directory and run:

```bash
go run ./ogrequotes/
```

# Acknowledgements

Thanks to the [`cretz/bine`](https://github.com/cretz/bine/) project for making using
Tor with Go a breeze. Some of the the examples use it to connect to Tor.
