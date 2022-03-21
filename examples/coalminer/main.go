package main

import (
	"context"
	"fmt"
	"time"

	"github.com/innix/shrek"
)

func main() {
	const addrsToFind = 5

	fmt.Println("This app will find", addrsToFind, ".onion addrs that start with 'coal'.")
	fmt.Println("Time to mine some coal... (press Ctrl+C to stop the program)")
	fmt.Println()
	time.Sleep(time.Second)

	coalMatcher := shrek.StartEndMatcher{
		Start: []byte("coal"),
	}

	for i := 0; i < addrsToFind; i++ {
		fmt.Print("Mining coal lump #", i+1, "... ")

		addr, err := shrek.MineOnionHostName(context.Background(), nil, coalMatcher)
		if err != nil {
			fmt.Println("mining failed, moving to next one...")
			fmt.Println("The error was:", err)

			continue
		}

		fmt.Println("got it:", addr.HostNameString())
	}

	fmt.Println()
	fmt.Println("All coal lumps have been mined!")
}
