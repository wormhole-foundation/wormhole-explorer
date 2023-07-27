package main

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/schollz/progressbar/v3"
	"github.com/wormhole-foundation/wormhole-explorer/common/client/alert"
	"github.com/wormhole-foundation/wormhole-explorer/common/dbhelpers"
	"github.com/wormhole-foundation/wormhole-explorer/fly/internal/metrics"
	"github.com/wormhole-foundation/wormhole-explorer/fly/storage"
	"go.uber.org/zap"
)

type GenericWorker func(ctx context.Context, repo *storage.Repository, item string) error

type Workpool struct {
	Workers    int
	Queue      chan string
	WG         sync.WaitGroup
	DB         *dbhelpers.Session
	Log        *zap.Logger
	Bar        *progressbar.ProgressBar
	WorkerFunc GenericWorker
}

type WorkerConfiguration struct {
	MongoURI      string `env:"MONGODB_URI,required"`
	MongoDatabase string `env:"MONGODB_DATABASE,required"`
	Filename      string `env:"FILENAME,required"`
	WorkerCount   int    `env:"WORKER_COUNT"`
}

func NewWorkpool(ctx context.Context, cfg WorkerConfiguration, workerFunc GenericWorker) *Workpool {

	wp := Workpool{
		Workers:    cfg.WorkerCount,
		Queue:      make(chan string, cfg.WorkerCount*1000),
		WG:         sync.WaitGroup{},
		Log:        zap.NewExample(),
		WorkerFunc: workerFunc,
	}

	db, err := dbhelpers.Connect(ctx, wp.Log, cfg.MongoURI, cfg.MongoDatabase)
	if err != nil {
		panic(err)
	}

	wp.DB = db

	for i := 0; i < cfg.WorkerCount; i++ {
		go wp.Process(ctx)
	}

	wp.WG.Add(cfg.WorkerCount)

	return &wp
}

func (w *Workpool) Process(ctx context.Context) error {
	repo := storage.NewRepository(alert.NewDummyClient(), metrics.NewDummyMetrics(), w.DB.Database, w.Log)
	var err error

	defer w.DB.DisconnectWithTimeout(10 * time.Second)

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
