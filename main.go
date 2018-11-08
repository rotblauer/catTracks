package main

import (
	"flag"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strconv"
	"sync"

	"github.com/rotblauer/catTracks/catTracks"
	"github.com/rotblauer/tileTester2/undump"
)

//start the url handlers, special init for everything?

//Toodle to do , Command line port arg, might mover er to main
func main() {
	var porty int
	var clearDBTestes bool
	var testesRun bool
	var buildIndexes bool
	var forwardurl string
	var tracksjsongzpath, tracksjsongzpathdevop, tracksjsongzpathedge string
	var dbpath, devopdbpath, edgedbpath string
	var dbpathtiles, devopdbpathtiles, edgedbpathtiles string
	var masterlock, devlock, edgelock string

	var procmaster, procedge bool

	flag.IntVar(&porty, "port", 8080, "port to serve and protect")
	flag.BoolVar(&clearDBTestes, "castrate-first", false, "clear out db of testes prefixed points") //TODO clear only certain values, ie prefixed with testes based on testesRun
	flag.BoolVar(&testesRun, "testes", false, "testes run prefixes name with testes-")              //hope that's your phone's name
	flag.BoolVar(&buildIndexes, "build-indexes", false, "build index buckets for original trackpoints")

	flag.StringVar(&forwardurl, "forward-url", "", "forward populate POST requests to this endpoint")

	flag.StringVar(&tracksjsongzpath, "tracks-gz-path", "", "path to appendable json.gz tracks (used by tippe)")
	flag.StringVar(&tracksjsongzpathdevop, "devop-gz-path", "", "path to appendable json.gz tracks (used by tippe) - for devop tipping")
	flag.StringVar(&tracksjsongzpathedge, "edge-gz-path", "", "path to appendable json.gz tracks (used by tippe) - for edge tipping")

	flag.StringVar(&dbpath, "db-path-master", path.Join("db", "tracks.db"), "path to master tracks bolty db")
	// these don't go to a bolt db, just straight to .json.gz
	flag.StringVar(&devopdbpath, "db-path-devop", "", "path to master tracks bolty db")
	flag.StringVar(&edgedbpath, "db-path-edge", "", "path to edge tracks bolty db")

	flag.StringVar(&dbpathtiles, "tiles-db-path-master", "", "path to master tiles tracks bolty db")
	flag.StringVar(&devopdbpathtiles, "tiles-db-path-devop", "", "path to master tiles tracks bolty db")
	flag.StringVar(&edgedbpathtiles, "tiles-db-path-edge", "", "path to edge tiles tracks bolty db")

	flag.StringVar(&masterlock, "master-lock", "", "path to master db lock")
	flag.StringVar(&devlock, "devop-lock", "", "path to devop db lock")
	flag.StringVar(&edgelock, "edge-lock", "", "path to edge db lock")

	flag.BoolVar(&procmaster, "proc-master", false, "run getem for master tiles")
	flag.BoolVar(&procedge, "proc-edge", false, "run getem for edge tiles")

	flag.Parse()

	catTracks.SetForwardPopulate(forwardurl)
	catTracks.SetLiveTracksGZ(tracksjsongzpath)
	catTracks.SetLiveTracksGZDevop(tracksjsongzpathdevop)
	catTracks.SetLiveTracksGZEdge(tracksjsongzpathedge)
	catTracks.SetDBPath("master", dbpath)
	catTracks.SetDBPath("devop", devopdbpath)
	catTracks.SetDBPath("edge", edgedbpath)

	catTracks.SetMasterLock(masterlock)
	catTracks.SetDevopLock(devlock)
	catTracks.SetEdgeLock(edgelock)

	// mkdir -p db/tracks.db
	// os.MkdirAll(filepath.Dir(edgedbpath), 0666)

	// Open Bolt DB.
	// catTracks.InitBoltDB()
	if bolterr := catTracks.InitBoltDB(); bolterr == nil {
		defer catTracks.GetDB("master").Close()
	}
	if clearDBTestes {
		e := catTracks.DeleteTestes()
		if e != nil {
			log.Println(e)
		}
	}
	if buildIndexes {
		catTracks.BuildIndexBuckets() //cleverly always returns nil
	}
	// if qterr := catTracks.InitQT(); qterr != nil {
	// 	log.Println("Error initing QT.")
	// 	log.Println(qterr)
	// }
	catTracks.InitMelody()
	catTracks.SetTestes(testesRun) //is false defaulter, false prefixes names with ""

	router := catTracks.NewRouter()

	http.Handle("/", router)

	var quitChan = make(chan bool)
	var mu sync.Mutex
	if procmaster {
		go func() {
			for {
				select {
				case <-quitChan:
					return
				default:
					// cat append all finished edge files to master.json.gz
					mu.Lock()
					b, err := ioutil.ReadFile(tracksjsongzpathedge)
					if err != nil {
						if os.IsNotExist(err) {
							os.Create(tracksjsongzpathedge)
							mu.Unlock()
							continue
						}
						log.Fatalln("procmaster/err:", err)
					}
					mu.Unlock()
					if len(b) == 0 {
						// log.Println("procmaster/nonerr", "continue")
						continue
					}
					fi, fe := os.OpenFile(tracksjsongzpath, os.O_WRONLY|os.O_APPEND, 0660)
					if fe != nil {
						if os.IsNotExist(fe) {
							os.Create(tracksjsongzpath) // should be only for dev
							log.Println("procmaster/err:", "created tracksjsongzpath")
							continue
						}
						panic(fe.Error())
					}
					if _, e := fi.Write(b); e != nil {
						panic(e)
					}
					fi.Close()

					mu.Lock()
					os.Truncate(tracksjsongzpathedge, 0)

					// move tracks-edge.db (mbtiles in bolty) -> tracks-devop.db
					os.Rename(tracksjsongzpathedge, tracksjsongzpathdevop)
					os.Rename(edgedbpathtiles, devopdbpathtiles)
					mu.Unlock()
					http.Get("http://localhost:8080/devop/refresh")

					// run tippe and undump on master
					// again, output should be to wip file, then mv
					// runTippe(out, in string, tilesetname string, bolttilesout string)
					out := filepath.Join(filepath.Dir(dbpath), "out.wip")
					in := tracksjsongzpath
					bolttilesout := filepath.Join(filepath.Dir(dbpath), "master-tiles.wip.db")
					if err := runTippe(out, in, "catTrack", bolttilesout); err != nil {
						panic(err.Error())
						// log.Println("TIPPERR master db tipp err:", err)
						// return
					}

					// os.Rename(out+".json.gz", filepath.Join(filepath.Dir(dbpath), "master.json.gz"))
					mu.Lock()
					os.Rename(bolttilesout, dbpathtiles)
					mu.Unlock()
					http.Get("http://localhost:8080/master/refresh")
					os.Remove(out + ".mbtiles")
				}
			}
		}()
	}

	if procedge {
		go func() {
			for {
				select {
				case <-quitChan:
					return
				case <-catTracks.NotifyNewEdge:
					mu.Lock()
					d := filepath.Dir(tracksjsongzpathedge)
					matches, err := filepath.Glob(filepath.Join(d, "*-fin-*"))
					if err != nil {
						panic("bad glob pattern:" + err.Error())
					}
					log.Printf("procedge matchesN=%d", len(matches))
					if len(matches) == 0 {
						mu.Unlock()
						continue
					}
					// cat and append all -fin- edges to edge.json.gz
					for _, m := range matches {
						b, err := ioutil.ReadFile(m)
						if err != nil {
							log.Println("err:", err)
							continue
						}
						fi, fe := os.OpenFile(tracksjsongzpathedge, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0660)
						if fe != nil {
							log.Println("fe:", fe)
							if fi != nil {
								fi.Close()
							}
							continue
						}
						_, e := fi.Write(b)
						fi.Close()
						if e != nil {
							log.Println("errappend:", e)
							continue
						}
						os.Remove(m)
					}
					// run tippe, note that this should lockmu and copy edge.json.gz to .snap
					// make a copy of edge.json.gz to edge.snap.json.gz
					b, e := ioutil.ReadFile(tracksjsongzpathedge)
					if e != nil {
						if os.IsNotExist(e) {
							os.Create(tracksjsongzpathedge)
							mu.Unlock()
							continue
						}
						panic(e)
					}
					snapEdgePath := filepath.Join(filepath.Dir(tracksjsongzpathedge), "edge.snap.json.gz")
					if e := ioutil.WriteFile(snapEdgePath, b, 0660); e != nil {
						panic(e)
					}
					mu.Unlock()
					err = runTippe(filepath.Join(filepath.Dir(tracksjsongzpathedge), "edge.out"), snapEdgePath, "catTrackEdge", filepath.Join(filepath.Dir(tracksjsongzpathedge), "edge-tiles.wip.db"))
					if err != nil {
						log.Println("TIPPERR:", err)
						continue
					}
					os.Remove(snapEdgePath)
					log.Println("waiting for lock ege for migrating")
					mu.Lock()
					log.Println("got lOCK")
					os.Rename(filepath.Join(filepath.Dir(tracksjsongzpathedge), "edge-tiles.wip.db"), edgedbpathtiles)
					os.Remove(filepath.Join(filepath.Dir(tracksjsongzpathedge), "edge.out.mbtiles"))
					// send req to tileserver to refresh edge db
					log.Println("sending refresh edge db get")
					mu.Unlock()
					http.Get("http://localhost:8080/edge/refresh")
					log.Println("finished refresh edge db get")
				}
			}
		}()
	}

	http.ListenAndServe(":"+strconv.Itoa(porty), nil)
	quitChan <- true
	quitChan <- true
}

func dotiproc() {
}

func runTippe(out, in string, tilesetname string, bolttilesout string) error {
	tippCmd, tippargs, tipperr := getTippyProcess(out, in, tilesetname)
	if tipperr != nil {
		return tipperr
	}

	log.Println("> [", tilesetname, "]", tippCmd, tippargs)
	tippmycanoe := exec.Command(tippCmd, tippargs...)
	tippmycanoe.Stdout = os.Stdout
	tippmycanoe.Stderr = os.Stderr

	err := tippmycanoe.Start()
	if err != nil {
		log.Println("Error starting Cmd", err)
		os.Exit(1)
	}

	if err := tippmycanoe.Wait(); err != nil {
		return err
	}

	log.Println("Dump: Migrating .mbtiles file back into a bolt db: ", bolttilesout)

	absoluteOut, err := filepath.Abs(out + ".mbtiles")
	if err != nil {
		log.Printf("err: %v", err)
		return err
	}
	undump.MbtilesToBolt(absoluteOut, bolttilesout)
	return nil
}

func getTippyProcess(out string, in string, tilesetname string) (tippCmd string, tippargs []string, err error) {
	//tippy process
	//Mapping extremely dense point data with vector tiles
	//https://www.mapbox.com/blog/vector-density/
	//-z19 -d11 -g3
	//"--no-tile-size-limit"
	//-as or --drop-densest-as-needed: If a tile is too large, try to reduce it to under 500K by increasing the minimum spacing between features. The discovered spacing applies to the entire zoom level.
	//-ag or --calculate-feature-density: Add a new attribute, tippecanoe_feature_density, to each feature, to record how densely features are spaced in that area of the tile. You can use this attribute in the style to produce a glowing effect where points are densely packed. It can range from 0 in the sparsest areas to 255 in the densest.
	//-pk or --no-tile-size-limit: Don't limit tiles to 500K bytes
	//-pf or --no-feature-limit: Don't limit tiles to 200,000 features
	//-pd or --force-feature-limit: Dynamically drop some fraction of features from large tiles to keep them under the 500K size limit. It will probably look ugly at the tile boundaries. (This is like -ad but applies to each tile individually, not to the entire zoom level.) You probably don't want to use this.
	//-r rate or --drop-rate=rate: Rate at which dots are dropped at zoom levels below basezoom (default 2.5). If you use -rg, it will guess a drop rate that will keep at most 50,000 features in the densest tile. You can also specify a marker-width with -rgwidth to allow fewer features in the densest tile to compensate for the larger marker, or -rfnumber to allow at most number features in the densest tile.
	//-z zoom or --maximum-zoom=zoom: Don't copy tiles from higher zoom levels than the specified zoom
	//-g gamma or --gamma=_gamma_: Rate at which especially dense dots are dropped (default 0, for no effect). A gamma of 2 reduces the number of dots less than a pixel apart to the square root of their original number.
	//-n name or --name=name: Set the tileset name
	//-ao or --reorder: Reorder features to put ones with the same properties in sequence, to try to get them to coalesce. You probably want to use this if you use --coalesce.
	//-aC or --cluster-densest-as-needed: If a tile is too large, try to reduce its size by increasing the minimum spacing between features, and leaving one placeholder feature from each group. The remaining feature will be given a "cluster": true attribute to indicate that it represents a cluster, a "point_count" attribute to indicate the number of features that were clustered into it, and a "sqrt_point_count" attribute to indicate the relative width of a feature to represent the cluster. If
	//- the features being clustered are points, the representative feature will be located at the average of the original points' locations; otherwise, one of the original features will be left as the representative
	//-M bytes or --maximum-tile-bytes=bytes: Use the specified number of bytes as the maximum compressed tile size instead of 500K.
	//-O features or --maximum-tile-features=features: Use the specified number of features as the maximum in a tile instead of 200,000.
	//-f or --force: Delete the mbtiles file if it already exists instead of giving an error
	//
	//WARNINGS:
	//Highest supported zoom with detail 14 is 18

	tippCmd = "/usr/local/bin/tippecanoe"
	tippargs = []string{
		"-ag",
		"-M", "1000000",
		"-O", "200000",
		"--cluster-densest-as-needed",
		"-g", "0.1",
		"--full-detail", "14",
		"--minimum-detail", "12",
		"-rg",
		"-rf100000",
		"--minimum-zoom", "3",
		"--maximum-zoom", "20",
		"-l", tilesetname, // TODO: what's difference layer vs name?
		"-n", tilesetname,
		"-o", out + ".mbtiles",
		"--force", "-P", in, "--reorder",
	}

	// 'in' should be an existing file
	_, err = os.Stat(in)
	if err != nil {
		return
	}

	// Use alternate tippecanoe path if 'bash -c which tippecanoe' returns something without error and different than default
	if b, e := exec.Command("bash -c", "which", "tippecanoe").Output(); e == nil && string(b) != tippCmd {
		tippCmd = string(b)
	}
	return
}
