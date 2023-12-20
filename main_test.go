package main

import (
	"log"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestBashExecLog(t *testing.T) {
	err := bashExec("echo 'hello world'", "abc ")
	if err != nil {
		t.Fatal(err)
	}
}

func TestFMR(t *testing.T) {

	testdir := filepath.Join(os.TempDir(), "test-fmr")
	defer os.RemoveAll(testdir)
	os.MkdirAll(testdir, 0755)

	fmr := newFileModRecorder(filepath.Join(testdir, "*.txt"))
	os.WriteFile(filepath.Join(testdir, "a.txt"), []byte("hello"), 0644)
	os.WriteFile(filepath.Join(testdir, "b.txt"), []byte("world"), 0644)
	fmr.record()
	time.Sleep(500 * time.Millisecond)
	os.WriteFile(filepath.Join(testdir, "a.txt"), []byte("helloooo"), 0644)
	fmr.mark()
	updated := fmr.updated()
	if len(updated) != 1 {
		t.Error("expected 1 updated files", len(updated))
	}
	for _, f := range fmr.files {
		t.Log(f.String())
	}
	t.Log(updated)
}

func TestProcmasterCHanPattern(t *testing.T) {
	masterCh := make(chan bool, 1)
	masterCh <- true
	defer close(masterCh)

	runMaster := func() {
		t.Log("running master...")
		time.Sleep(5 * time.Second)
	}

	go func() {
		for {
			select {
			case <-masterCh:
				runMaster()
			}
		}
	}()

	for i := 0; i < 15; i++ {
		time.Sleep(1 * time.Second)
		log.Println("running edge...")
		select {
		case masterCh <- true:
			log.Println("masterCh <- true")
		default:
			log.Println("master is busy")
		}
	}
}
