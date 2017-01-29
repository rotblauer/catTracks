package catTracks

// // DeleteSpain deletes spain
// func DeleteSpain() error {
// 	e := GetDB().Update(func(tx *bolt.Tx) error {
// 		b := tx.Bucket([]byte("tracks"))
// 		c := b.Cursor()
// 		for k, v := c.First(); k != nil; k, v = c.Next() {
// 			var tp trackPoint.TrackPoint
// 			e := json.Unmarshal(v, &tp)
// 			if e != nil {
// 				fmt.Println("Error deleting testes.")
// 				return e
// 			}
// 			if tp.Lng < 12.0 && tp.Lng > -10.0 {
// 				b.Delete(k)
// 			}
// 		}
// 		return nil
// 	})
// 	return e
// }
