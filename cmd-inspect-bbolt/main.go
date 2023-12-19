package main

import (
	"flag"
	"log"

	bolt "go.etcd.io/bbolt"
)

var snapBucketKey = []byte("catsnaps")

var flagSourceDB = flag.String("source", "catTracks-old.db", "source database")

func main() {
	flag.Parse()

	// open source db
	source, err := bolt.Open(*flagSourceDB, 0666, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer source.Close()
	log.Println("source", *flagSourceDB)

	// copy snaps
	err = source.View(func(stx *bolt.Tx) error {
		b := stx.Bucket(snapBucketKey)
		if b == nil {
			log.Fatalln("no snaps bucket in source db")
		}
		b.ForEach(func(k, v []byte) error {
			log.Println("key", string(k), "value", string(v))
			return nil
		})
		return nil
	})
}
