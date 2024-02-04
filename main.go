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
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/paulmach/orb/geojson"

	"github.com/rotblauer/catTrackslib"
)

var flagExportPostGIS = flag.Bool("flagExportPostGIS", false, "export to postgis")

// Toodle to do , Command line port arg, might mover er to main
func main() {
	var flagPort int
	var flagclearDBTestes bool
	var flagTestesRun bool
	// var flagBuildIndexes bool
	var flagForwardURL string
	var flagTracksjsongzpathMaster, flagTracksjsongzpathDevop, flagTracksjsongzpathEdge string
	var flagDBPathMaster, flagDBDevopPath, flagDBPathEdge string
	var flagMasterLock, flagDevLock, flagEdgeLock string

	var placesLayer bool

	var flagProcmster, flagProcedge bool
	var flagTippeEdgeMaxSeconds int64

	var flagPostGISExportTarget string // target postgis endpoint

	flag.IntVar(&flagPort, "port", 8080, "port to serve and protect")
	flag.BoolVar(&flagclearDBTestes, "castrate-first", false, "clear out db of testes prefixed points") // TODO clear only certain values, ie prefixed with testes based on testesRun
	flag.BoolVar(&flagTestesRun, "testes", false, "testes run prefixes name with testes-")              // hope that's your phone's name
	// flag.BoolVar(&flagBuildIndexes, "build-indexes", false, "build index buckets for original trackpoints")

	flag.StringVar(&flagForwardURL, "forward-url", "", "forward populate POST requests to this endpoint")

	flag.StringVar(&flagTracksjsongzpathMaster, "tracks-gz-path", "", "path to appendable json.gz tracks (used by tippe)")
	flag.StringVar(&flagTracksjsongzpathDevop, "devop-gz-path", "", "path to appendable json.gz tracks (used by tippe) - for devop tipping")
	flag.StringVar(&flagTracksjsongzpathEdge, "edge-gz-path", "", "path to appendable json.gz tracks (used by tippe) - for edge tipping")

	flag.StringVar(&flagDBPathMaster, "db-path-master", path.Join("db", "tracks.db"), "path to master tracks bolty db")
	// these don't go to a bolt db, just straight to .json.gz
	flag.StringVar(&flagDBDevopPath, "db-path-devop", "", "path to master tracks bolty db")
	flag.StringVar(&flagDBPathEdge, "db-path-edge", "", "path to edge tracks bolty db")

	flag.StringVar(&flagMasterLock, "master-lock", "", "path to master db lock")
	flag.StringVar(&flagDevLock, "devop-lock", "", "path to devop db lock")
	flag.StringVar(&flagEdgeLock, "edge-lock", "", "path to edge db lock")

	flag.BoolVar(&flagProcmster, "proc-master", false, "run getem for master tiles")
	flag.BoolVar(&flagProcedge, "proc-edge", false, "run getem for edge tiles")
	flag.Int64Var(&flagTippeEdgeMaxSeconds, "tippe-edge-max-seconds", 60, "max seconds to run tippe on edge tiles")

	flag.BoolVar(&placesLayer, "places-layer", false, "generate layer for valid ios places")

	flag.StringVar(&flagPostGISExportTarget, "export.target", "postgres://postgres:mysecretpassword@localhost:5432/cattracks1?sslmode=prefer", "target postgis endpoint")

	flag.Parse()

	catTrackslib.SetForwardPopulate(flagForwardURL)
	catTrackslib.SetLiveTracksGZ(flagTracksjsongzpathMaster)
	catTrackslib.SetLiveTracksGZDevop(flagTracksjsongzpathDevop)
	catTrackslib.SetLiveTracksGZEdge(flagTracksjsongzpathEdge)
	catTrackslib.SetDBPath("master", flagDBPathMaster)
	catTrackslib.SetDBPath("devop", flagDBDevopPath)
	catTrackslib.SetDBPath("edge", flagDBPathEdge)

	catTrackslib.SetMasterLock(flagMasterLock)
	catTrackslib.SetDevopLock(flagDevLock)
	catTrackslib.SetEdgeLock(flagEdgeLock)

	catTrackslib.SetPlacesLayer(placesLayer)

	// mkdir -p db/tracks.db
	// os.MkdirAll(filepath.Dir(edgedbpath), 0666)

	// Open Bolt DB.
	// catTrackslib.InitBoltDB()
	if bolterr := catTrackslib.InitBoltDB(); bolterr == nil {
		defer catTrackslib.GetDB("master").Close()
	}
	if flagclearDBTestes {
		e := catTrackslib.DeleteTestes()
		if e != nil {
			log.Println(e)
		}
	}
	// if flagBuildIndexes {
	// 	catTrackslib.BuildIndexBuckets() // cleverly always returns nil
	// }
	// if qterr := catTrackslib.InitQT(); qterr != nil {
	// 	log.Println("Error initing QT.")
	// 	log.Println(qterr)
	// }

	if flagExportPostGIS != nil && *flagExportPostGIS {
		log.Println("Exporting PostGIS")
		catTrackslib.ExportPostGIS(flagPostGISExportTarget)
		return
	}

	// FIXME: This is deprecated/dilapidated because
	// we don't actually use websockets for anything.
	// But if we did, this would be a way and place to start
	// hacking something in there.
	catTrackslib.InitMelody()

	// Defaults false, causing names prefixed with: ""
	// Apparently configures a test environment.
	catTrackslib.SetTestes(flagTestesRun)

	// Does boilerplate for setting up the router.
	// Configures routes, which are defined in routes.go.
	router := catTrackslib.NewRouter()
	http.Handle("/", router)

	// These are our always-on workers.
	// They are go routines running `tippecanoe` commands
	// to generate .mbtiles (mapbox tiles) map tiles.
	// The 'master' routine generates .mbtiles for each cats' tracks.
	// The 'master' routine ingests the latest tracks from the 'edge' routine, and then truncates that data.
	// The 'edge' routines generates tiles for only the latest tracks (everyone included).
	var quitChan = make(chan bool)
	var edgeMutex sync.Mutex

	splitCatCellsOutputRoot := filepath.Join(filepath.Dir(flagDBPathMaster), "cat-cells")
	splitCatCellsDBRoot := filepath.Join(splitCatCellsOutputRoot, "dbs")
	genMBTilesPath := filepath.Join(splitCatCellsOutputRoot, "mbtiles")

	// tilesetsDir is where consbio/mbtileserver -d serves from, with --enable-fs-watch on.
	tilesetsDir := filepath.Join(filepath.Dir(flagDBPathMaster), "tilesets")
	os.MkdirAll(tilesetsDir, 0755)

	procMasterPrefixed := func(label string) string {
		if label == "" {
			return "[procmaster] "
		}
		return fmt.Sprintf("[procmaster: %s] ", label)
	}

	// procmasterCh is a channel which is used to trigger the procmaster routine.
	// The edge routine will trigger the procmaster routine when it has finished processing the latest tracks.
	procmasterCh := make(chan bool, 1)
	defer close(procmasterCh)

	// procmasterCh <- true // init procmaster run

	if flagProcmster {
		go func() {
		procmasterloop:
			for {
				select {
				case <-quitChan:
					return
				case <-procmasterCh:
					log.Println("[procmaster] starting iter")

					// tileRecovery:
					// - if spurious .mbtiles-journals files exist (tippecanoe interrupted)
					// - if no target .mbtiles directory exists (first run)
					var tileRecovery = false
					if _, err := os.Stat(genMBTilesPath); os.IsNotExist(err) {
						log.Printf("[procmaster] WARN: no genMBTilesPath found: %s, recovering\n", genMBTilesPath)
						tileRecovery = true
					}

					// recover corrupted mbtiles in case something got fucked, like killed or something
					// any .mbtiles-journal files in the genMBTilesDir are considered indicators of corrupted .mbtiles files
					// if there, we delete both the .mbtiles and .mbtiles-journals files, and touch the corresponding .json.gz file to update modtime, allowing a rebuild
					if !tileRecovery {
						mbtilesJournals, _ := filepath.Glob(filepath.Join(genMBTilesPath, "*.mbtiles-journal"))
						if len(mbtilesJournals) > 0 {
						} else {
							log.Println("[procmaster] zero .mbtiles-journal files found, no recovery needed")
						}
						for _, journalFilepath := range mbtilesJournals {
							if _, err := os.Stat(journalFilepath); os.IsNotExist(err) {
								// probably impossible
							} else if err == nil {
								// this is actually what we don't want; if any .mbtiles-journal exist, the corresponding .mbtiles file should be considered corrupted
								tileRecovery = true
								tilesFilepath := strings.ReplaceAll(journalFilepath, ".mbtiles-journal", ".mbtiles")
								log.Println("[procmaster] WARN: found ", journalFilepath, ", considering corrupted: ", tilesFilepath)
								_ = bashExec(fmt.Sprintf("rm %s", journalFilepath), procMasterPrefixed(""))
								_ = bashExec(fmt.Sprintf("rm %s", tilesFilepath), procMasterPrefixed(""))
							}
						}
					}

					// edgeSize is the initial size of the edge file.
					// we don't need the lock here because we only care about its minimum size, and mutating
					// processes are still locked.
					edgeSize := int64(0)
					if fi, err := os.Stat(flagTracksjsongzpathEdge); err == nil {
						edgeSize = fi.Size()
					} else if !tileRecovery {
						log.Println("procmaster: edge.json.gz errored, skipping (sleep 1m)", err)
						// time.Sleep(time.Minute)
						continue procmasterloop
					}

					if edgeSize < 100 && !tileRecovery {
						log.Printf("procmaster: edge.json.gz is too small (%d < 100 bytes), skipping (sleep 1m)\n", edgeSize)
						// time.Sleep(time.Minute)
						continue procmasterloop
					} else {
						log.Printf("procmaster: edge.json.gz is %d bytes, running\n", edgeSize)
					}

					// handle migrating init run
					if _, err := os.Stat(splitCatCellsOutputRoot); os.IsNotExist(err) {
						// run command to split master to cat.json.gz by unique cells
						// will mkdir -p required output and db dirs
						// eg.
						//   ~/tdata/cat-cells/{ia,rye}.json.gz
						//   ~/tdata/cat-cells/dbs/{ia,rye}.db
						if err := runCatCellSplitter(flagTracksjsongzpathMaster, splitCatCellsOutputRoot, splitCatCellsDBRoot); err != nil {
							log.Fatalln(err)
						}
					}

					// split the edge into cats
					if edgeSize > 100 {
						// now the master -> cat.json.gz is split into cells
						// so we can run the edge -> cat.json.gz
						edgeMutex.Lock()

						if err := runCatCellSplitter(flagTracksjsongzpathEdge, splitCatCellsOutputRoot, splitCatCellsDBRoot); err != nil {
							log.Fatalln(err)
						}

						// cat append all finished edge files to master.json.gz
						// append edge tracks to master
						_ = bashExec(fmt.Sprintf("cat %s >> %s", flagTracksjsongzpathEdge, flagTracksjsongzpathMaster), procMasterPrefixed(""))

						log.Println("rolling edge to develop")
						// rename edge.json.gz -> devop.json.gz (roll)
						log.Println(procMasterPrefixed(""), fmt.Sprintf("mv %s %s", flagTracksjsongzpathEdge, flagTracksjsongzpathDevop))
						_ = os.Rename(flagTracksjsongzpathEdge, flagTracksjsongzpathDevop)

						// touch edge.json.gz
						log.Println(procMasterPrefixed(""), fmt.Sprintf("touch %s", flagTracksjsongzpathEdge))
						_ = os.Truncate(flagTracksjsongzpathEdge, 0)

						// _, _ = os.Create(tracksjsongzpathEdge) // create or truncate
						// rename tilesets/edge.mbtiles ->  tilesets/devop.mbtiles (roll)
						log.Println(procMasterPrefixed(""), fmt.Sprintf("mv %s %s", filepath.Join(tilesetsDir, "edge.mbtiles"), filepath.Join(tilesetsDir, "devop.mbtiles")))
						_ = os.Rename(filepath.Join(tilesetsDir, "edge.mbtiles"), filepath.Join(tilesetsDir, "devop.mbtiles"))

						edgeMutex.Unlock()
					}

					// // did the cattracks-split-cats-uniqcell-gz command generate any new tracks?
					// // or were they all dupes?
					// // if they were all dupes, we can skip the rest of this procmaster iter
					// fmrJSONGZs.mark()
					// if len(fmrJSONGZs.updated()) == 0 && mbTilesExist {
					// 	log.Println("[procmaster] cat-cells/*.json.gz unmodified, short-circuiting")
					// 	continue procmasterloop
					// }

					// run tippe on all cat cells .json.gzs.
					// eg.
					//  ~/tdata/cat-cells/mbtiles
					var fmrMBTiles = newFileModRecorder(filepath.Join(genMBTilesPath, "*.mbtiles"))

					fmrMBTiles.record()
					_ = bashExec(fmt.Sprintf(`time tippecanoe-walk-dir --source %s --output %s`, splitCatCellsOutputRoot, genMBTilesPath), procMasterPrefixed("tippecanoe-walk-dir"))
					fmrMBTiles.mark()

					// if tippe on the tracks didn't change any mbtiles, we can skip the rest
					updatedMBTiles := fmrMBTiles.updated()
					if len(updatedMBTiles) == 0 {
						log.Println("[procmaster] cat-cells/*.mbtiles unmodified, short-circuiting")
						continue procmasterloop
					}

					// copy changed files individually to save time,
					// at the expense of being a more brittle approach.
					// copying all files takes about 40 seconds
					/*
						root@rottor:~/go/src/github.com/rotblauer/catTracks# du -sh ~/tdata/cat-cells/mbtiles
						6.4G    /root/tdata/cat-cells/mbtiles
					*/
					// _ = bashExec(fmt.Sprintf("time cp %s/*.mbtiles %s/", genMBTilesPath, tilesetsDir), procMasterPrefixed(""))
					for _, u := range updatedMBTiles {
						_ = bashExec(fmt.Sprintf("time cp %s %s/", u, tilesetsDir), procMasterPrefixed(""))
					}

					// genpop cats long naps low lats
					//
					// genpop.mbtiles will be the union of all .mbtiles for cats who are not ia or rye
					//
					// collect all .mbtiles for cats who are not ia or rye
					// then run tile-join on them to make genpop.mbtiles
					// genpop is expected to be much smaller than either ia or rye
					// only do this if any one of the genpop people have pushed tracks and have new tiles
					// 	// TODO
					// 	// problems: need to skip genpop.mbtiles,
					// 	// and exclude cats from genpop with scapegoat algorithms
					genPopTilesPath := filepath.Join(genMBTilesPath, "genpop.level-23.mbtiles")
					genPopTilesExist := false
					if _, err := os.Stat(genPopTilesPath); err == nil {
						genPopTilesExist = true
					}

					genPopCatMBTiles := []string{} // will be all tile paths EXCEPT those matching any of notGenPop
					notGenPop := []string{
						"ia",
						"rye",
					}
				genpoploop:
					for _, u := range updatedMBTiles {
						// path/to/ia.level-23.mbtiles => ia
						// path/to/bob.mbtiles => bob
						for _, reservedName := range notGenPop {
							if strings.Contains(strings.Split(filepath.Base(u), ".")[0], reservedName) {
								continue genpoploop
							}
						}
						genPopCatMBTiles = append(genPopCatMBTiles, u)
					}

					if len(genPopCatMBTiles) == 0 && genPopTilesExist {
						log.Println("[procmaster] genpop tiles not updated, skipping tile-join and cp .mbtiles")
						continue procmasterloop
					}
					log.Println("[procmaster] genpop tiles updated, running tile-join")

					// run tile-join on them to make genpop.mbtiles
					genPopTilePathsString := strings.Join(genPopCatMBTiles, " ")
					_ = bashExec(fmt.Sprintf("time tile-join --force --no-tile-size-limit -o %s %s", genPopTilesPath, genPopTilePathsString), procMasterPrefixed("tile-join"))

					// TODO we have now TWO copies of relatively fresh mbtiles dirs,
					// we need to keep the genMBTilesPath so we can avoid re-genning stale json.gz->tiles,
					// and we need to keep tilesets/ clean so we can avoid trouble (.mbtiles-journals) with the mbtiles server
					// good news is these .mbtiles dbs are relatively small, < 10GB
					// Copy the newly-generated (or updated) .mbtiles files to the tilesets/ dir which gets served.
					// Expect live-reload (consbio/mbtileserver --enable-fs-watch) to pick them up.
					// cp ~/tdata/cat-cells/mbtiles/*.mbtiles ~/tdata/tilesets/
					_ = bashExec(fmt.Sprintf("time cp %s %s/", genPopTilesPath, tilesetsDir), procMasterPrefixed(""))

					log.Println("[procmaster] finished iter")
				}
			}
		}()
	}

	if flagProcmster && flagProcedge {
		go func() {
			debounceFireProcMaster := false
			for {
				select {
				case <-quitChan:
					return
				case <-catTrackslib.NotifyNewEdge:

					// drain buffered chan, just in case
					for len(catTrackslib.NotifyNewEdge) > 0 {
						<-catTrackslib.NotifyNewEdge
					}

					procEdgePrefix := "[procedge]"

					// this function processes the edge.json.gz file.
					// these are tracks which have not yet been included in master.json.gz.
					// this function is only allowed to append to edge.json.gz file, and to copy it;
					// it should NEVER delete or truncate the edge.json.file, that is the job of procmaster
					// when it ingests the edge.json.gz file into master.json.gz.

					log.Printf("%sstarting iter\n", procEdgePrefix)
					rootDir := filepath.Dir(flagTracksjsongzpathEdge)

					// if no fin files, then this is a re-run trigger after a short interval from previous.
					// channel can be like that.
					if matches, _ := filepath.Glob(filepath.Join(rootDir, "*-fin-*")); len(matches) == 0 {
						log.Printf("%sno fin files, skipping\n", procEdgePrefix)
						continue
					}

					// lock the edge file, competing with prcmaster
					edgeMutex.Lock()

					directMasterGZPath := filepath.Join(rootDir, "direct-"+filepath.Base(flagTracksjsongzpathMaster))
					if _, err := os.Stat(directMasterGZPath); err != nil {
						// copy master.json.gz to direct-master.json.gz
						log.Println("[procedge] copying master.json.gz to direct-master.json.gz")
						_ = bashExec(fmt.Sprintf("cp %s %s", flagTracksjsongzpathMaster, directMasterGZPath), procEdgePrefix)
					}

					// look for any _fin_ished partial edge files, and dump them into edge.json.gz
					_ = bashExec(fmt.Sprintf("cat %s/*-fin-* >> %s", rootDir, flagTracksjsongzpathEdge), procEdgePrefix)
					// then dump these new tracks directly to direct-master.json.gz
					_ = bashExec(fmt.Sprintf("cat %s/*-fin-* >> %s", rootDir, directMasterGZPath), procEdgePrefix)

					// then remove all _fin_ished partial edge files
					log.Println(procEdgePrefix, fmt.Sprintf("rm %s/*-fin-*", rootDir))
					fins, _ := filepath.Glob(filepath.Join(rootDir, "*-fin-*"))
					for _, fin := range fins {
						os.Remove(fin)
					}

					// copy edge.json.gz to edge.snap.json.gz, for use as a snapshot with tippecanoe
					snapEdgePath := filepath.Join(rootDir, "edge.snap.json.gz")
					_ = bashExec(fmt.Sprintf("cp %s %s", flagTracksjsongzpathEdge, snapEdgePath), procEdgePrefix)

					edgeMutex.Unlock()

					// run tippecanoe on the snapshotted edge data

					tippeStart := time.Now()

					var err error
					err = runTippe(filepath.Join(rootDir, "edge.mbtiles"), snapEdgePath, "catTrackEdge")
					if err != nil {
						log.Println("[procedge] tippecanoe errored:", err)
						continue
					}

					tippeTook := time.Since(tippeStart)
					log.Printf("[procedge] tippecanoe took %s\n", tippeTook.Round(time.Microsecond))

					// remove the snapshot after use
					log.Println(procEdgePrefix, fmt.Sprintf("rm %s", snapEdgePath))
					os.Remove(snapEdgePath)

					log.Println("[procedge] waiting for lock ege for migrating")
					edgeMutex.Lock()
					log.Println("[procedge] got lock")
					// move the new edge mbtiles to the tilesets dir for serving
					log.Println(procEdgePrefix, fmt.Sprintf("mv %s %s",
						filepath.Join(rootDir, "edge.mbtiles"), filepath.Join(tilesetsDir, "edge.mbtiles")))
					os.Rename(filepath.Join(rootDir, "edge.mbtiles"), filepath.Join(tilesetsDir, "edge.mbtiles"))
					edgeMutex.Unlock()

					log.Println("[procedge] finished iter")

					if tippeTook < time.Second*time.Duration(flagTippeEdgeMaxSeconds) {
						log.Printf("[procedge] tippecanoe took less than %d seconds, skipping procmaster trigger\n", flagTippeEdgeMaxSeconds)
						continue
					}

					// debounce for when many tracks are posted in batches,
					// so there'll be many of these fired in succession,
					// and we want to trigger master for the latest of these
					go func() {
						if debounceFireProcMaster {
							return
						}
						debounceFireProcMaster = true
						defer func() {
							debounceFireProcMaster = false
						}()
						time.Sleep(60 * time.Second)
						select {
						case procmasterCh <- true:
							log.Println("[procedge] procmasterCh <- true")
						default:
							log.Println("[procedge] procmasterCh full, skipping")
						}
					}()
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
						baseDataDir := filepath.Dir(flagTracksjsongzpathEdge)
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

	http.ListenAndServe(":"+strconv.Itoa(flagPort), nil)
	quitChan <- true
	quitChan <- true
	quitChan <- true
}

type prefixedWriter struct {
	*log.Logger
}

func (pw prefixedWriter) Write(p []byte) (n int, err error) {
	pw.Logger.Println(string(p))
	return len(p), nil
}

func bashExec(cmd, logPrefix string) error {
	if strings.LastIndex(logPrefix, " ") != len(logPrefix)-1 {
		logPrefix = logPrefix + " "
	}
	log.Printf("%sbash: %s\n", logPrefix, cmd)
	bashCmd := exec.Command("bash", "-c", cmd)
	stdout := prefixedWriter{log.New(os.Stdout, logPrefix, log.LstdFlags|log.Lmsgprefix)}
	bashCmd.Stdout = stdout
	stderr := prefixedWriter{log.New(os.Stderr, logPrefix, log.LstdFlags|log.Lmsgprefix)}
	bashCmd.Stderr = stderr
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

	prefix := fmt.Sprintf("[%s] ", tilesetname)
	stdout := prefixedWriter{log.New(os.Stdout, prefix, log.LstdFlags|log.Lmsgprefix)}
	tippmycanoe.Stdout = stdout
	stderr := prefixedWriter{log.New(os.Stderr, prefix, log.LstdFlags|log.Lmsgprefix)}
	tippmycanoe.Stderr = stderr

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
		// "-EElevation:max",
		// "-ESpeed:max", // mean",
		// "-EAccuracy:mean",
		// // "-EActivity:concat", // might get huge
		// "-EPressure:mean",
		// "--include", "UnixTime",
		"--include", "UUID",
		"--include", "Name",
		"--include", "Activity",
		"--include", "Elevation",
		"--include", "Speed",
		"--include", "Accuracy",
		// "--include", "Heading",
		"--include", "UnixTime",
		"-EUnixTime:max",
		"-EElevation:max",
		"-ESpeed:max", // mean",
		"-EAccuracy:mean",
		"--single-precision",
		"-r1", // == --drop-rate
		// "-rg",
		// "-rf100000",
		// "-g", "2",
		// "--full-detail", "12",
		// "--minimum-detail", "12",
		"--minimum-zoom", "3",
		"--maximum-zoom", "18",
		"--json-progress", "--progress-interval", "30",
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

type fileMod struct {
	fpath     string
	modBefore time.Time // will be 0 if newly created
	modAfter  time.Time // will be 0 if deleted
}

func (f fileMod) String() string {
	return fmt.Sprintf("%s: %s -> %s", f.fpath, f.modBefore, f.modAfter)
}

type fileModRecorder struct {
	glob  string
	files []fileMod
}

func newFileModRecorder(glob string) *fileModRecorder {
	return &fileModRecorder{
		glob: glob,
	}
}

func (fmr *fileModRecorder) record() error {
	fmr.files = []fileMod{}
	matches, _ := filepath.Glob(fmr.glob)
	for _, match := range matches {
		fi, err := os.Stat(match)
		if err != nil {
			return err
		}
		fmr.files = append(fmr.files, fileMod{
			fpath:     match,
			modBefore: fi.ModTime(),
		})
	}
	return nil
}

func (fmr *fileModRecorder) mark() error {
	matches, _ := filepath.Glob(fmr.glob)
	for _, match := range matches {
		fi, err := os.Stat(match)
		if err != nil {
			return err
		}
		matchedExisting := false
		for i, f := range fmr.files {
			if f.fpath == match {
				// updated file
				matchedExisting = true
				fmr.files[i].modAfter = fi.ModTime()
				break
			}
		}
		// created file
		if !matchedExisting {
			fmr.files = append(fmr.files, fileMod{
				fpath:    match,
				modAfter: fi.ModTime(),
			})
		}
	}
	// // deleted files ; UNNECESSARY because zero-value of modAfter is time.Zero
	// for _, f := range fmr.files {
	// 	_, err := os.Stat(f.fpath)
	// 	if os.IsNotExist(err) {
	// 		f.modAfter = time.Time{}
	// 	} else {
	// 		return err
	// 	}
	// }
	return nil
}

// updated returns all updated filepaths in the order of most recently updated first
func (fmr *fileModRecorder) updated() []string {
	ret := []string{}
	sort.Slice(fmr.files, func(i, j int) bool {
		return fmr.files[i].modAfter.After(fmr.files[j].modAfter)
	})
	for _, f := range fmr.files {
		updated := f.modAfter.After(f.modBefore)
		deleted := !f.modAfter.IsZero() && f.modAfter.IsZero()
		created := f.modBefore.IsZero() && !f.modAfter.IsZero()
		if (created || updated) && !deleted {
			ret = append(ret, f.fpath)
		}
	}
	return ret
}
