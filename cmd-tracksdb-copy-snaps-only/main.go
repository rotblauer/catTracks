package main

import (
	"flag"
	"log"
	bolt "go.etcd.io/bbolt"
)

var flagSourceDB = flag.String("source", "catTracks-old.db", "source database")
var flagTargetDB = flag.String("target", "catTracks-new.db", "target database")

func main() {
	flag.Parse()

	log.Println("source", *flagSourceDB)
	log.Println("target", *flagTargetDB)

	// open source db
	_, err := bolt.Open(*flagSourceDB, 0666, nil)
	if err != nil {
		log.Fatal(err)
	}
	// defer source.Close()

}
