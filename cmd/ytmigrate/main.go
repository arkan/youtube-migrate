package main

import (
	"log"

	// "github.com/julienschmidt/httprouter"

	yt_migrate "github.com/arkan/yt-migrate"
)

func main() {
	log.Printf("Please authenticate with old account")
	clFrom, err := yt_migrate.New()
	if err != nil {
		panic(err)
	}
	subs, err := clFrom.GetSubscriptions()
	if err != nil {
		panic(err)
	}

	log.Printf("Please authenticate with new account")
	clTo, err := yt_migrate.New()
	if err != nil {
		panic(err)
	}
	for _, s := range subs {
		if err := clTo.AddSubscription(s); err != nil {
			log.Fatalf("Unable to add subscription: %s\n", err.Error())
		}
	}

	log.Printf("Migration completed")
}
