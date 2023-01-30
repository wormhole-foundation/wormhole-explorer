package main

import (
	"bufio"
	"bytes"
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"github.com/briandowns/spinner"
	"github.com/schollz/progressbar/v3"
)

type Backfiller struct {
	Filename string
	Strategy string
	Workpool *Workpool
}

func (b *Backfiller) Run() error {
	f, err := os.Open(b.Filename)
	if err != nil {
		return err
	}

	s := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
	s.Color("red")
	s.Suffix = fmt.Sprintf(" counting lines")

	s.Start()
	pLines, err := b.countLines()
	if err != nil {
		return err
	}
	s.Stop()

	fmt.Printf("lines: %d \n ", pLines)

	b.Workpool.Bar = progressbar.Default(int64(pLines))

	counter := 0
	defer f.Close()

	r := bufio.NewReader(f)

	// read file line by line and send to workpool
	for {
		line, _, err := r.ReadLine() //loading chunk into buffer
		if err != nil {
			if err == io.EOF {
				break
			}
			log.Fatalf("a real error happened here: %v\n", err)
		}
		b.Workpool.Queue <- string(line)
		counter += 1
	}

	// send exit signal to all workers
	for i := 0; i < b.Workpool.Workers; i++ {
		b.Workpool.Queue <- "exit"
	}

	// wait for all workers to finish
	b.Workpool.WG.Wait()

	fmt.Printf("processed %d lines\n", counter)

	return nil
}

func (b *Backfiller) countLines() (int, error) {
	file, _ := os.Open(b.Filename)
	defer file.Close()

	buf := make([]byte, 32*1024)
	count := 0
	lineSep := []byte{'\n'}

	for {
		c, err := file.Read(buf)
		count += bytes.Count(buf[:c], lineSep)

		switch {
		case err == io.EOF:
			return count, nil

		case err != nil:
			return count, err
		}
	}

}

func main() {

	var filename string
	var strategy string
	flag.StringVar(&filename, "file", "", "file to process (mandatory)")
	flag.StringVar(&strategy, "strategy", "vaa", "strategy to use (vaa|txhash)")

	flag.Parse()

	if os.Getenv("MONGODB_URI") == "" {
		os.Setenv("MONGODB_URI", "mongodb://localhost:27017/")
		fmt.Println("MONGODB_URI not set, using default")
	}

	if filename == "" {
		flag.Usage()
		os.Exit(1)
	}

	ctx := context.Background()

	// choose strategy
	var worker GenericWorker

	switch strategy {
	case "vaa":
		worker = workerVaa
	case "txhash":
		worker = workerTxHash
	default:
		flag.Usage()
		os.Exit(1)
	}

	wp := NewWorkpool(ctx, 100, worker)

	b := Backfiller{
		Filename: filename,
		Workpool: wp,
	}
	err := b.Run()
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println("done!")
}
