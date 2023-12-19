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
		sourceBucket := stx.Bucket(snapBucketKey)
		if sourceBucket == nil {
			log.Fatalln("no snaps bucket in source db")
		}
		sourceBucket.ForEach(func(k, v []byte) error {
			updateErr := target.Update(func(ttx *bolt.Tx) error {
				targetBucket, targetErr := ttx.CreateBucketIfNotExists(snapBucketKey)
				if targetErr != nil {
					return targetErr
				}
				log.Println("copying", string(k))
				targetErr = targetBucket.Put(k, v)
				return targetErr
			})
			if updateErr != nil {
				return updateErr
			}
			return nil
		})

		return nil
	})
	if err != nil {
		log.Fatal(err)
	}
}
