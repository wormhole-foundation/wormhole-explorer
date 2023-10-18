package main

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/schollz/progressbar/v3"
	"github.com/wormhole-foundation/wormhole-explorer/common/dbutil"
	"github.com/wormhole-foundation/wormhole-explorer/fly/storage"
	"go.uber.org/zap"
)

type GenericWorker func(ctx context.Context, repo *storage.Repository, item string) error

type Workpool struct {
	Workers    int
	Queue      chan string
	WG         sync.WaitGroup
	DB         *dbutil.Session
	Log        *zap.Logger
	Bar        *progressbar.ProgressBar
	WorkerFunc GenericWorker
	Repository *storage.Repository
}

type WorkerConfiguration struct {
	MongoURI       string `env:"MONGODB_URI,required"`
	MongoDatabase  string `env:"MONGODB_DATABASE,required"`
	Filename       string `env:"FILENAME,required"`
	WorkerCount    int    `env:"WORKER_COUNT"`
	NotifyEnabled  bool   `env:"NOTIFY_ENABLED"`
	AwsRegion      string `env:"AWS_REGION"`
	AwsAccessKeyId string `env:"AWS_ACCESS_KEY_ID"`
	AwsSecretKey   string `env:"AWS_SECRET_ACCESS_KEY"`
	AwsEndpoint    string `env:"AWS_ENDPOINT"`
	AwsSnsURL      string `env:"AWS_SNS_URL"`
}

func NewWorkpool(ctx context.Context, cfg WorkerConfiguration, workerFunc GenericWorker) *Workpool {

	wp := Workpool{
		Workers:    cfg.WorkerCount,
		Queue:      make(chan string, cfg.WorkerCount*1000),
		WG:         sync.WaitGroup{},
		Log:        zap.NewExample(),
		WorkerFunc: workerFunc,
	}

	db, err := dbutil.Connect(ctx, wp.Log, cfg.MongoURI, cfg.MongoDatabase, false)
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
	var err error

	defer w.DB.DisconnectWithTimeout(10 * time.Second)

	for {
		select {
		case line := <-w.Queue:
			if line == "exit" {
				w.WG.Done()
				return nil
			}
			err = w.WorkerFunc(ctx, w.Repository, line)
			if err != nil {
				fmt.Println(err)
				break
			}

			w.Bar.Add(1) // its safe to call Add concurrently

		}
	}

}
