package main

import (
	"context"
	"fmt"
	"os"
	"sync"

	"github.com/schollz/progressbar/v3"
	"github.com/wormhole-foundation/wormhole-explorer/fly/storage"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
)

type GenericWorker func(ctx context.Context, repo *storage.Repository, item string) error

type Workpool struct {
	Workers    int
	Queue      chan string
	WG         sync.WaitGroup
	DB         *mongo.Database
	Log        *zap.Logger
	Bar        *progressbar.ProgressBar
	WorkerFunc GenericWorker
}

func NewWorkpool(ctx context.Context, workers int, workerFunc GenericWorker) *Workpool {

	wp := Workpool{
		Workers:    workers,
		Queue:      make(chan string, workers*1000),
		WG:         sync.WaitGroup{},
		Log:        zap.NewExample(),
		WorkerFunc: workerFunc,
	}

	db, err := storage.GetDB(ctx, wp.Log, os.Getenv("MONGODB_URI"), "wormhole")
	if err != nil {
		panic(err)
	}

	wp.DB = db

	for i := 0; i < workers; i++ {
		go wp.Process(ctx)
	}

	wp.WG.Add(workers)

	return &wp
}

func (w *Workpool) Process(ctx context.Context) error {
	repo := storage.NewRepository(w.DB, w.Log)
	var err error

	for {
		select {
		case line := <-w.Queue:
			if line == "exit" {
				w.WG.Done()
				return nil
			}
			err = w.WorkerFunc(ctx, repo, line)
			if err != nil {
				fmt.Println(err)
				break
			}

			w.Bar.Add(1) // its safe to call Add concurrently

		}
	}

}
