package main

import (
	"flag"
	"log"

	bolt "go.etcd.io/bbolt"
)

var snapBucketKey = []byte("catsnaps")

var flagSourceDB = flag.String("source", "catTracks-old.db", "source database")
var flagTargetDB = flag.String("target", "catTracks-new.db", "target database")

func main() {
	flag.Parse()

	// open source db
	source, err := bolt.Open(*flagSourceDB, 0666, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer source.Close()
	log.Println("source", *flagSourceDB)

	// open target db
	target, err := bolt.Open(*flagTargetDB, 0666, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer target.Close()
	log.Println("target", *flagTargetDB)

	// copy snaps
	err = source.View(func(stx *bolt.Tx) error {
		b := stx.Bucket(snapBucketKey)
		if b == nil {
			log.Fatalln("no snaps bucket in source db")
		}
		c := b.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {

			err := target.Update(func(ttx *bolt.Tx) error {
				bb, err := ttx.CreateBucketIfNotExists(snapBucketKey)
				if err != nil {
					return err
				}
				log.Println("copying", string(k))
				err = bb.Put(k, v)
				return err
			})
			if err != nil {
				return err
			}
		}
		return nil
	})
}
