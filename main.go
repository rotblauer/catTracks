package main

import (
	"compress/gzip"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/kpawlik/geojson"

	"github.com/rotblauer/catTrackslib"
)

var exportPostGIS = flag.Bool("exportPostGIS", false, "export to postgis")

// Toodle to do , Command line port arg, might mover er to main
func main() {
	var porty int
	var clearDBTestes bool
	var testesRun bool
	var buildIndexes bool
	var forwardurl string
	var tracksjsongzpathMaster, tracksjsongzpathDevop, tracksjsongzpathEdge string
	var dbpath, devopdbpath, edgedbpath string
	var masterlock, devlock, edgelock string

	var placesLayer bool

	var procmaster, procedge bool

	var exportTarget string // target postgis endpoint

	flag.IntVar(&porty, "port", 8080, "port to serve and protect")
	flag.BoolVar(&clearDBTestes, "castrate-first", false, "clear out db of testes prefixed points") // TODO clear only certain values, ie prefixed with testes based on testesRun
	flag.BoolVar(&testesRun, "testes", false, "testes run prefixes name with testes-")              // hope that's your phone's name
	flag.BoolVar(&buildIndexes, "build-indexes", false, "build index buckets for original trackpoints")

	flag.StringVar(&forwardurl, "forward-url", "", "forward populate POST requests to this endpoint")

	flag.StringVar(&tracksjsongzpathMaster, "tracks-gz-path", "", "path to appendable json.gz tracks (used by tippe)")
	flag.StringVar(&tracksjsongzpathDevop, "devop-gz-path", "", "path to appendable json.gz tracks (used by tippe) - for devop tipping")
	flag.StringVar(&tracksjsongzpathEdge, "edge-gz-path", "", "path to appendable json.gz tracks (used by tippe) - for edge tipping")

	flag.StringVar(&dbpath, "db-path-master", path.Join("db", "tracks.db"), "path to master tracks bolty db")
	// these don't go to a bolt db, just straight to .json.gz
	flag.StringVar(&devopdbpath, "db-path-devop", "", "path to master tracks bolty db")
	flag.StringVar(&edgedbpath, "db-path-edge", "", "path to edge tracks bolty db")

	flag.StringVar(&masterlock, "master-lock", "", "path to master db lock")
	flag.StringVar(&devlock, "devop-lock", "", "path to devop db lock")
	flag.StringVar(&edgelock, "edge-lock", "", "path to edge db lock")

	flag.BoolVar(&procmaster, "proc-master", false, "run getem for master tiles")
	flag.BoolVar(&procedge, "proc-edge", false, "run getem for edge tiles")
	flag.BoolVar(&placesLayer, "places-layer", false, "generate layer for valid ios places")

	flag.StringVar(&exportTarget, "export.target", "postgres://postgres:mysecretpassword@localhost:5432/cattracks1?sslmode=prefer", "target postgis endpoint")

	flag.Parse()

	catTrackslib.SetForwardPopulate(forwardurl)
	catTrackslib.SetLiveTracksGZ(tracksjsongzpathMaster)
	catTrackslib.SetLiveTracksGZDevop(tracksjsongzpathDevop)
	catTrackslib.SetLiveTracksGZEdge(tracksjsongzpathEdge)
	catTrackslib.SetDBPath("master", dbpath)
	catTrackslib.SetDBPath("devop", devopdbpath)
	catTrackslib.SetDBPath("edge", edgedbpath)

	catTrackslib.SetMasterLock(masterlock)
	catTrackslib.SetDevopLock(devlock)
	catTrackslib.SetEdgeLock(edgelock)

	catTrackslib.SetPlacesLayer(placesLayer)

	// mkdir -p db/tracks.db
	// os.MkdirAll(filepath.Dir(edgedbpath), 0666)

	// Open Bolt DB.
	// catTrackslib.InitBoltDB()
	if bolterr := catTrackslib.InitBoltDB(); bolterr == nil {
		defer catTrackslib.GetDB("master").Close()
	}
	if clearDBTestes {
		e := catTrackslib.DeleteTestes()
		if e != nil {
			log.Println(e)
		}
	}
	if buildIndexes {
		catTrackslib.BuildIndexBuckets() // cleverly always returns nil
	}
	// if qterr := catTrackslib.InitQT(); qterr != nil {
	// 	log.Println("Error initing QT.")
	// 	log.Println(qterr)
	// }

	if exportPostGIS != nil && *exportPostGIS {
		log.Println("Exporting PostGIS")
		catTrackslib.ExportPostGIS(exportTarget)
		return
	}

	// FIXME: This is deprecated/dilapidated because
	// we don't actually use websockets for anything.
	// But if we did, this would be a way and place to start
	// hacking something in there.
	catTrackslib.InitMelody()

	// Defaults false, causing names prefixed with: ""
	// Apparently configures a test environment.
	catTrackslib.SetTestes(testesRun)

	// Does boilerplate for setting up the router.
	// Configures routes, which are defined in routes.go.
	router := catTrackslib.NewRouter()
	http.Handle("/", router)

	// These are our always-on workers.
	// They are go routines running `tippecanoe` commands
	// to generate .mbtiles (mapbox tiles) map tiles.
	// The 'master' routine generates map tiles for all cat tracks
	// for the whole world. It takes a long time, around 24 hours.
	// The 'edge' routines generates tiles for the latest
	// cat tracks.
	// IIRC, the master routine loop truncates the edge tracks
	// list, so there is a period of time where (while master runs) some (eg. yesterday's) tracks
	// are not expected to be shown on tiles.
	var quitChan = make(chan bool)
	var edgeMutex sync.Mutex

	splitCatCellsOutputRoot := filepath.Join(filepath.Dir(dbpath), "cat-cells")
	splitCatCellsDBRoot := filepath.Join(splitCatCellsOutputRoot, "dbs")
	tilesetsDir := filepath.Join(filepath.Dir(dbpath), "tilesets")
	os.MkdirAll(tilesetsDir, 0755)

	procMasterPrefixed := func(label string) string {
		return fmt.Sprintf("[proc-master: %s] ", label)
	}

	var _catsJSONGZLastModTime = time.Time{}

	if procmaster {
		go func() {
		procmasterloop:
			for {
				select {
				case <-quitChan:
					return
				default:
					log.Println("[procmaster] starting iter")

					if fi, err := os.Stat(tracksjsongzpathEdge); err == nil {
						if fi.Size() < 100 {
							log.Println("procmaster: edge.json.gz is too small, skipping")
							time.Sleep(time.Minute)
							continue
						}
					} else if err != nil {
						log.Println("procmaster: edge.json.gz errored, skipping", err)
						time.Sleep(time.Minute)
						continue
					} else {
						log.Println("procmaster: edge.json.gz is %d bytes, running", fi.Size())
					}

					// cat append all finished edge files to master.json.gz

					// handle migrating init run
					if _, err := os.Stat(splitCatCellsOutputRoot); os.IsNotExist(err) {
						// run command to split master to cat.json.gz by unique cells
						// will mkdir -p required output and db dirs
						// eg.
						//   ~/tdata/cat-cells/{ia,rye}.json.gz
						//   ~/tdata/cat-cells/dbs/{ia,rye}.db
						if err := runCatCellSplitter(tracksjsongzpathMaster, splitCatCellsOutputRoot, splitCatCellsDBRoot); err != nil {
							log.Fatalln(err)
						}
					}
					// now the master -> cat.json.gz is split into cells
					// so we can run the edge -> cat.json.gz
					edgeMutex.Lock()

					if err := runCatCellSplitter(tracksjsongzpathEdge, splitCatCellsOutputRoot, splitCatCellsDBRoot); err != nil {
						log.Fatalln(err)
					}

					// append edge tracks to master
					_ = bashExec(fmt.Sprintf("time cat %s >> %s", tracksjsongzpathEdge, tracksjsongzpathMaster), "")

					log.Println("rolling edge to develop")
					// rename edge.json.gz -> devop.json.gz (roll)
					_ = os.Rename(tracksjsongzpathEdge, tracksjsongzpathDevop)
					// touch edge.json.gz
					_, _ = os.Create(tracksjsongzpathEdge) // create or truncate
					// rename tilesets/edge.mbtiles ->  tilesets/devop.mbtiles (roll)
					_ = os.Rename(filepath.Join(filepath.Dir(dbpath), "tilesets", "edge.mbtiles"), filepath.Join(filepath.Dir(dbpath), "tilesets", "devop.mbtiles"))

					edgeMutex.Unlock()

					// did the cattracks-split-cats-uniqcell-gz command generate any new .mbtiles?
					// or were they all dupes?
					// if they were all dupes, we can skip the rest of this procmaster iter
					// TODO
					catsGZMatches, err := filepath.Glob(filepath.Join(splitCatCellsOutputRoot, "*.json.gz"))
					if err != nil {
						log.Fatalln(err)
					}
					if len(catsGZMatches) > 0 {
						catsJSONGZLastModTime := time.Time{}
						for _, catGZ := range catsGZMatches {
							if fi, err := os.Stat(catGZ); err == nil {
								if fi.ModTime().After(catsJSONGZLastModTime) {
									catsJSONGZLastModTime = fi.ModTime()
								}
							}
						}
						if !catsJSONGZLastModTime.After(_catsJSONGZLastModTime) {
							log.Println("[procmaster] cat-cells/*.json.gz unmodified, short-circuiting")
							continue procmasterloop
						}
						_catsJSONGZLastModTime = catsJSONGZLastModTime
					}

					// run tippe on split cat cells
					// eg.
					//  ~/tdata/cat-cells/mbtiles
					genMBTilesPath := filepath.Join(splitCatCellsOutputRoot, "mbtiles")
					_ = bashExec(fmt.Sprintf(`time tippecanoe-walk-dir --source %s --output %s`, splitCatCellsOutputRoot, genMBTilesPath), procMasterPrefixed("tippecanoe-walk-dir"))

					_ = bashExec(fmt.Sprintf("time cp %s/*.mbtiles %s/", genMBTilesPath, tilesetsDir), "")

					// genpop cats long naps low lats
					//
					// genpop.mbtiles will be the union of all .mbtiles for cats who are not ia or rye
					//
					// collect all .mbtiles for cats who are not ia or rye
					// then run tile-join on them to make genpop.mbtiles
					// genpop is expected to be much smaller than either ia or rye
					// only do this if any one of the genpop people have pushed tracks and have new tiles
					genPopTilesPath := filepath.Join(genMBTilesPath, "genpop.mbtiles")

					// get the modtime of genpop.tiles
					// we'll use this to compare to the modtime of all the .mbtiles.
					// stale .mbtiles belonging to genpop cats will be tile-joined with genpop.mbtiles
					var genPopTilesModTime time.Time
					if fi, err := os.Stat(genPopTilesPath); err == nil {
						genPopTilesModTime = fi.ModTime()
					}

					// genPopDidUpdate tells us if any of the .mbtiles for the genpop has updated more
					// recently than the modtime on genpop.mbtiles
					var genPopDidUpdate bool

					genPopTilePaths := []string{} // will be all tile paths EXCEPT those matching any of notGenPop
					notGenPop := []string{
						"ia",
						"rye",
					}

					// TODO
					// problems: need to skip genpop.mbtiles,
					// and exclude cats from genpop with scapegoat algorithms
					//
					// isGenPopException := func(fi os.FileInfo) bool {
					// 	// path/to/ia.level-23.mbtiles => ia
					// 	// path/to/bob.mbtiles => bob
					// 	// for _, reservedName := range notGenPop {
					// 	// 	if strings.Contains(strings.Split(filepath.Base(fi.Name()), ".")[0], reservedName) {
					// 	// 		return true
					// 	// 	}
					// 	// }
					// 	return fi.Size() > 100000000 // 100MB (100000000)
					// 	return false
					// }

					// get the glob list of all generated .mbtiles
					allpopTiles, err := filepath.Glob(filepath.Join(genMBTilesPath, "*.mbtiles"))
					if err != nil {
						log.Fatalln(err)
					}

				genPopLoop:
					for _, tilesFile := range allpopTiles {
						for _, reservedName := range notGenPop {
							// path/to/ia.level-23.mbtiles => ia
							// path/to/bob.mbtiles => bob
							if strings.Contains(strings.Split(filepath.Base(tilesFile), ".")[0], reservedName) {
								continue genPopLoop
							}
						}

						// no hit; is unreserved genpop mbtiles
						genPopTilePaths = append(genPopTilePaths, tilesFile)

						// mark if this mbtiles has been update more recently than genpop.mbtiles
						if fi, err := os.Stat(tilesFile); err == nil {
							if fi.ModTime().After(genPopTilesModTime) {
								genPopDidUpdate = true
							}
						}
					}

					if !genPopDidUpdate {
						log.Println("[procmaster] genpop tiles not updated, skipping tile-join and cp")
						continue procmasterloop
					}

					log.Println("genpop tiles updated, running tile-join")
					// run tile-join on them to make genpop.mbtiles
					genPopTilePathsString := strings.Join(genPopTilePaths, " ")
					_ = bashExec(fmt.Sprintf("time tile-join --force --no-tile-size-limit -o %s %s", genPopTilesPath, genPopTilePathsString), procMasterPrefixed("tile-join"))

					// TODO we have now TWO copies of relatively fresh mbtiles dirs,
					// we need to keep the genMBTilesPath so we can avoid re-genning stale json.gz->tiles,
					// and we need to keep tilesets/ clean so we can avoid trouble (.mbtiles-journals) with the mbtiles server
					// good news is these .mbtiles dbs are relatively small, < 10GB
					// Copy the newly-generated (or updated) .mbtiles files to the tilesets/ dir which gets served.
					// Expect live-reload (consbio/mbtileserver --enable-fs-watch) to pick them up.
					// cp ~/tdata/cat-cells/mbtiles/*.mbtiles ~/tdata/tilesets/
					_ = bashExec(fmt.Sprintf("time cp %s/*.mbtiles %s/", genMBTilesPath, tilesetsDir), "")

					log.Println("finished procmaster iter")

					// // run tippe and undump on master
					// // again, output should be to wip file, then mv
					// // runTippe(out, in string, tilesetname string, bolttilesout string)
					// out := filepath.Join(filepath.Dir(dbpath), "master.mbtiles")
					// in := tracksjsongzpathMaster
					// log.Println("running master tippe")
					// if err := runTippe(out, in, "catTrack"); err != nil {
					// 	panic(err.Error())
					// 	// log.Println("TIPPERR master db tipp err:", err)h
					// 	// return
					// }
					//
					// // splitterCmd.Stdin =
					//
					// // os.Rename(out+".json.gz", filepath.Join(filepath.Dir(dbpath), "master.json.gz"))
					//
					// os.Rename(out, filepath.Join(tilesetsDir, "master.mbtiles"))
					// os.Remove(out + ".mbtiles")
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
				case <-catTrackslib.NotifyNewEdge:

					log.Println("[procedge] starting iter")

					// look for any finished edge geojson gz files
					edgeMutex.Lock()
					d := filepath.Dir(tracksjsongzpathEdge)

					_ = bashExec(fmt.Sprintf("cat %s/*-fin-* >> %s", d, tracksjsongzpathEdge), "procedge: ")
					_ = bashExec(fmt.Sprintf("rm %s/*-fin-*", d), "procedge: ")

					snapEdgePath := filepath.Join(filepath.Dir(tracksjsongzpathEdge), "edge.snap.json.gz")
					_ = bashExec(fmt.Sprintf("cp %s %s", tracksjsongzpathEdge, snapEdgePath), "procedge: ")

					// matches, err := filepath.Glob(filepath.Join(d, "*-fin-*"))
					// if err != nil {
					// 	panic("bad glob pattern:" + err.Error())
					// }
					// log.Printf("procedge matchesN=%d", len(matches))
					// if len(matches) == 0 {
					// 	edgeMutex.Unlock()
					// 	continue
					// }
					//
					// // cat and append all -fin- edges to edge.json.gz
					// for _, m := range matches {
					// 	b, err := ioutil.ReadFile(m)
					// 	if err != nil {
					// 		log.Println("err:", err)
					// 		continue
					// 	}
					// 	fi, fe := os.OpenFile(tracksjsongzpathEdge, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0660)
					// 	if fe != nil {
					// 		log.Println("fe:", fe)
					// 		if fi != nil {
					// 			fi.Close()
					// 		}
					// 		continue
					// 	}
					// 	_, e := fi.Write(b)
					// 	fi.Close()
					// 	if e != nil {
					// 		log.Println("errappend:", e)
					// 		continue
					// 	}
					// 	os.Remove(m)
					// }
					// // run tippe, note that this should lockmu and copy edge.json.gz to .snap
					// // make a copy of edge.json.gz to edge.snap.json.gz
					// b, e := ioutil.ReadFile(tracksjsongzpathEdge)
					// if e != nil {
					// 	if os.IsNotExist(e) {
					// 		os.Create(tracksjsongzpathEdge)
					// 		edgeMutex.Unlock()
					// 		continue
					// 	}
					// 	panic(e)
					// }
					// if e := ioutil.WriteFile(snapEdgePath, b, 0660); e != nil {
					// 	panic(e)
					// }

					edgeMutex.Unlock()

					var err error
					err = runTippe(filepath.Join(d, "edge.mbtiles"), snapEdgePath, "catTrackEdge")
					if err != nil {
						log.Println("TIPPERR:", err)
						continue
					}
					os.Remove(snapEdgePath)
					log.Println("[procedge] waiting for lock ege for migrating")
					edgeMutex.Lock()
					log.Println("[procedge] got lock")
					os.Rename(filepath.Join(filepath.Dir(tracksjsongzpathEdge), "edge.mbtiles"), filepath.Join(tilesetsDir, "edge.mbtiles"))
					// os.Remove(filepath.Join(filepath.Dir(tracksjsongzpathedge), "edge.out.mbtiles"))
					// send req to tileserver to refresh edge db
					edgeMutex.Unlock()

					log.Println("[procedge] finished iter")
				}
			}
		}()
	}

	if placesLayer {
		go func() {
			var placesProcLock sync.Mutex
			places := []*geojson.Feature{}
			// Even though this runs on the /1sec range, it won't fire until the previous run has finished;
			for _ = range time.Tick(1 * time.Second) {
				select {
				case <-quitChan:
					return
				case p := <-catTrackslib.FeaturePlaceChan:
					places = append(places, p)
					log.Println("+place:", p)
				default:
					if lenp := len(places); lenp > 0 {
						log.Println("processing", lenp, "places: ", places)
						// eg. /var/tdata/places.json.gz
						baseDataDir := filepath.Dir(tracksjsongzpathEdge)
						placesJSONGZ := filepath.Join(baseDataDir, "places.json.gz")

						placesProcLock.Lock() // might be unnecessary

						pgz := catTrackslib.CreateGZ(placesJSONGZ, gzip.BestCompression)
						for _, f := range places {
							if f == nil {
								continue
							}
							pgz.JE().Encode(f)
						}
						catTrackslib.CloseGZ(pgz)

						// reset local places
						places = []*geojson.Feature{}

						wipTilesDB := filepath.Join(baseDataDir, "places.mbtiles")
						err := runTippeLite(wipTilesDB, placesJSONGZ, "catTrackPlace")
						if err != nil {
							log.Println("tippe/places/err:", err)
							placesProcLock.Unlock()
							continue
						}

						os.Rename(wipTilesDB, filepath.Join(baseDataDir, "tilesets", "places.mbtiles"))
						// os.Remove(filepath.Join(baseDataDir, "places.out.mbtiles"))

						placesProcLock.Unlock()

						log.Println("finished processing", lenp, "places")
					}
					continue
				}
			}
		}()
	}

	http.ListenAndServe(":"+strconv.Itoa(porty), nil)
	quitChan <- true
	quitChan <- true
	quitChan <- true
}

func bashExec(cmd, logPrefix string) error {
	log.Println("bash executing:", cmd)
	bashCmd := exec.Command("bash", "-c", cmd)
	bashCmd.Stdout = log.New(os.Stdout, logPrefix, log.LstdFlags|log.Lmsgprefix).Writer()
	bashCmd.Stderr = log.New(os.Stderr, logPrefix, log.LstdFlags|log.Lmsgprefix).Writer()
	return bashCmd.Run()
}

func runCatCellSplitter(sourceGZ, outputRoot, dbRoot string) error {
	/*
		time cat ~/tdata/master.json.gz | zcat |\
		    go run . \
		    --workers 8 \
		    --cell-level 23 \
		    --batch-size 100000 \
		    --cache-size 50000000 \
		    --compression-level 9
	*/
	c := fmt.Sprintf(`time zcat %s | cattracks-split-cats-uniqcell-gz \
--workers 4 \
--cell-level 23 \
--batch-size 100000 \
--cache-size 5000000 \
--compression-level 9 \
--output %s \
--db-root %s`,
		sourceGZ, outputRoot, dbRoot)

	return bashExec(c, "cattracks-split-cats-uniqcell-gz: ")
}

func runTippeLite(out, in string, tilesetname string) error {
	tippCmd, tippargs, tipperr := getTippyProcessLite(out, in, tilesetname)
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
	return nil
}

func runTippe(out, in string, tilesetname string) error {
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
	return nil
}

// doesn't do as much squashing
func getTippyProcessLite(out string, in string, tilesetname string) (tippCmd string, tippargs []string, err error) {
	// tippy process
	// Mapping extremely dense point data with vector tiles
	// https://www.mapbox.com/blog/vector-density/
	// WARNINGS:
	// Highest supported zoom with detail 14 is 18
	tippCmd = "/usr/local/bin/tippecanoe"
	tippargs = []string{
		// -ag or --calculate-feature-density: Add a new attribute, tippecanoe_feature_density, to each feature, to record how densely features are spaced in that area of the tile. You can use this attribute in the style to produce a glowing effect where points are densely packed. It can range from 0 in the sparsest areas to 255 in the densest.
		"-ag",
		// -M bytes or --maximum-tile-bytes=bytes: Use the specified number of bytes as the maximum compressed tile size instead of 500K.
		// "-M", "1000000",
		// -O features or --maximum-tile-features=features: Use the specified number of features as the maximum in a tile instead of 200,000.
		"-O", "200",
		// -aC or --cluster-densest-as-needed: If a tile is too large, try to reduce its size by increasing the minimum spacing between features, and leaving one placeholder feature from each group. The remaining feature will be given a "cluster": true attribute to indicate that it represents a cluster, a "point_count" attribute to indicate the number of features that were clustered into it, and a "sqrt_point_count" attribute to indicate the relative width of a feature to represent the cluster. If
		"--cluster-densest-as-needed",
		// -g gamma or --gamma=_gamma_: Rate at which especially dense dots are dropped (default 0, for no effect). A gamma of 2 reduces the number of dots less than a pixel apart to the square root of their original number.
		"-g", "0",
		// TODO: document.
		"--full-detail", "14",
		"--minimum-detail", "12",
		// -r rate or --drop-rate=rate: Rate at which dots are dropped at zoom levels below basezoom (default 2.5). If you use -rg, it will guess a drop rate that will keep at most 50,000 features in the densest tile. You can also specify a marker-width with -rgwidth to allow fewer features in the densest tile to compensate for the larger marker, or -rfnumber to allow at most number features in the densest tile.
		"-rg",
		"-rf1000",
		"--minimum-zoom", "3",
		// -z zoom or --maximum-zoom=zoom: Don't copy tiles from higher zoom levels than the specified zoom
		"--maximum-zoom", "18",
		"-l", tilesetname, // TODO: what's difference layer vs name?
		// -n name or --name=name: Set the tileset name
		"-n", tilesetname,
		"-o", out,
		// -f or --force: Delete the mbtiles file if it already exists instead of giving an error
		"--force",
		"-P", in,
		// -ao or --reorder: Reorder features to put ones with the same properties in sequence, to try to get them to coalesce. You probably want to use this if you use --coalesce.
		// "--reorder",
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

func getTippyProcess(out string, in string, tilesetname string) (tippCmd string, tippargs []string, err error) {
	// tippy process
	// Mapping extremely dense point data with vector tiles
	// https://www.mapbox.com/blog/vector-density/
	// -z19 -d11 -g3
	// "--no-tile-size-limit"
	// -as or --drop-densest-as-needed: If a tile is too large, try to reduce it to under 500K by increasing the minimum spacing between features. The discovered spacing applies to the entire zoom level.
	// -ag or --calculate-feature-density: Add a new attribute, tippecanoe_feature_density, to each feature, to record how densely features are spaced in that area of the tile. You can use this attribute in the style to produce a glowing effect where points are densely packed. It can range from 0 in the sparsest areas to 255 in the densest.
	// -pk or --no-tile-size-limit: Don't limit tiles to 500K bytes
	// -pf or --no-feature-limit: Don't limit tiles to 200,000 features
	// -pd or --force-feature-limit: Dynamically drop some fraction of features from large tiles to keep them under the 500K size limit. It will probably look ugly at the tile boundaries. (This is like -ad but applies to each tile individually, not to the entire zoom level.) You probably don't want to use this.
	// -r rate or --drop-rate=rate: Rate at which dots are dropped at zoom levels below basezoom (default 2.5). If you use -rg, it will guess a drop rate that will keep at most 50,000 features in the densest tile. You can also specify a marker-width with -rgwidth to allow fewer features in the densest tile to compensate for the larger marker, or -rfnumber to allow at most number features in the densest tile.
	// -z zoom or --maximum-zoom=zoom: Don't copy tiles from higher zoom levels than the specified zoom
	// -g gamma or --gamma=_gamma_: Rate at which especially dense dots are dropped (default 0, for no effect). A gamma of 2 reduces the number of dots less than a pixel apart to the square root of their original number.
	// -n name or --name=name: Set the tileset name
	// -ao or --reorder: Reorder features to put ones with the same properties in sequence, to try to get them to coalesce. You probably want to use this if you use --coalesce.
	// -aC or --cluster-densest-as-needed: If a tile is too large, try to reduce its size by increasing the minimum spacing between features, and leaving one placeholder feature from each group. The remaining feature will be given a "cluster": true attribute to indicate that it represents a cluster, a "point_count" attribute to indicate the number of features that were clustered into it, and a "sqrt_point_count" attribute to indicate the relative width of a feature to represent the cluster. If
	// - the features being clustered are points, the representative feature will be located at the average of the original points' locations; otherwise, one of the original features will be left as the representative
	// -M bytes or --maximum-tile-bytes=bytes: Use the specified number of bytes as the maximum compressed tile size instead of 500K.
	// -O features or --maximum-tile-features=features: Use the specified number of features as the maximum in a tile instead of 200,000.
	// -f or --force: Delete the mbtiles file if it already exists instead of giving an error
	//
	// WARNINGS:
	// Highest supported zoom with detail 14 is 18

	tippCmd = "/usr/local/bin/tippecanoe"
	// tilesFPBase := filepath.Join(filepath.Dir(out), "ttiles", "$1", "$2") // $1 and $2 are first 2 of 3 argument (Z, X) passed from tippe to arbitrary pre/post-processing shell cmd
	tippargs = []string{

		// ADD max tile bytes -> 300k? .. thinking dat must be a big slow; if ye have to download 10mb maps everytime load, no wonder slow
		"--maximum-tile-bytes", "330000", // num bytes/tile,default: 500kb=500000
		// "--maximum-tile-bytes", "250000", // num bytes/tile,default: 500kb=500000
		// "--maximum-tile-features", "200000", // num feats/tile,default=200000
		"--cluster-densest-as-needed",
		// "--cluster-distance", "2",
		"--cluster-distance=1",
		"--calculate-feature-density",
		// "-j", `{ "catTrack": [ "any", [">", "Speed", 0], ["!has", "Activity"] , [ "all", ["!=", "Activity", "Stationary"], ["!=", "Activity", "Unknown"] ] ] }`,
		// "-j", `{ "catTrack": [ "any", ["!has", "Accuracy"], ["<=", "Accuracy", 200], [ "<=", "$zoom", 13 ] ] }`, // NOT catTrackEdge; only take high-accuracy (<=11m) points for high-level (close up) zooms

		// -Eattribute:operation or --accumulate-attribute=attribute:operation: Preserve the named attribute from features that are dropped, coalesced-as-needed, or clustered. The operation may be sum, product, mean, max, min, concat, or comma to specify how the named attribute is accumulated onto the attribute of the same name in a feature that does survive, eg. --accumulate-attribute=POP_MAX:sum
		"-EElevation:max",
		"-ESpeed:max", // mean",
		"-EAccuracy:mean",
		// "-EActivity:concat", // might get huge
		"-EPressure:mean",
		"-r1", // == --drop-rate
		// "-rg",
		// "-rf100000",
		// "-g", "2",
		// "--full-detail", "12",
		// "--minimum-detail", "12",
		"--minimum-zoom", "3",
		"--maximum-zoom", "18",
		"-l", tilesetname, // TODO: what's difference layer vs name?
		"-n", tilesetname,
		"-o", out,
		"--force",
		"--read-parallel", in,
		// "--preserve-input-order",

		// -C 'mkdir -p tiles/$1/$2; tee tiles/$1/$2/$3.geojson'
		// "-c", fmt.Sprintf(`mkdir -p %s; tee %s`, tilesFPBase, filepath.Join(tilesFPBase, "$3.geojson")),

		// "--reorder",
		// "--no-progress-indicator",
		// "--version",

		// // "-g", "3", # running without gamma
		// // "--maximum-tile-bytes", "50000", // num bytes/tile,default: 500kb=500000
		// // "--maximum-tile-features", "200000", // num feats/tile,default=200000
		// "--cluster-densest-as-needed",
		// "--cluster-distance", "2",
		// "--calculate-feature-density",
		// "-rg",
		// // "-rf100000",
		// // "-g", "2",
		// "--full-detail", "14",
		// "--minimum-detail", "12",
		// "--minimum-zoom", "3",
		// "--maximum-zoom", "19",
		// "-l", tilesetname, // TODO: what's difference layer vs name?
		// "-n", tilesetname,
		// "-o", out + ".mbtiles",
		// "--force",
		// "--read-parallel", in,
		// "--preserve-input-order",
		// // "--reorder",
		// // "--no-progress-indicator",
		// // "--version",

		// // // R1:TIPPING dis mor
		// "-g", "3",
		// // "--maximum-tile-bytes", "50000", // num bytes/tile,default: 500kb=500000
		// // "--maximum-tile-features", "200000", // num feats/tile,default=200000
		// "--cluster-densest-as-needed",
		// // "--cluster-distance", "2",
		// "--calculate-feature-density",
		// "-rg",
		// // "-rf100000",
		// // "-g", "2",
		// "--full-detail", "14",
		// "--minimum-detail", "12",
		// "--minimum-zoom", "3",
		// "--maximum-zoom", "18",
		// "-l", tilesetname, // TODO: what's difference layer vs name?
		// "-n", tilesetname,
		// "-o", out + ".mbtiles",
		// "--force",
		// "--read-parallel", in,
		// // "--preserve-input-order",
		// "--reorder",
		// // "--no-progress-indicator",
		// // "--version",

		// "-ag",
		// "-M", "1000000",
		// "-O", "200000",
		// "--cluster-densest-as-needed",
		// "-g", "0.1",
		// "--full-detail", "14",
		// "--minimum-detail", "12",
		// "-rg",
		// "-rf100000",
		// "--minimum-zoom", "3",
		// "--maximum-zoom", "20",
		// "-l", tilesetname, // TODO: what's difference layer vs name?
		// "-n", tilesetname,
		// "-o", out + ".mbtiles",
		// "--force",
		// "-P", in,
		// "--reorder",
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
