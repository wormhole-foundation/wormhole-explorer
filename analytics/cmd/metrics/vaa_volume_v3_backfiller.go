package metrics

import (
	"context"
	"encoding/binary"
	"fmt"
	"github.com/wormhole-foundation/wormhole-explorer/analytics/cmd/token"
	"github.com/wormhole-foundation/wormhole-explorer/analytics/metric"
	"github.com/wormhole-foundation/wormhole-explorer/analytics/prices"
	"github.com/wormhole-foundation/wormhole-explorer/common/client/parser"
	"github.com/wormhole-foundation/wormhole-explorer/common/domain"
	"github.com/wormhole-foundation/wormhole-explorer/common/logger"
	"github.com/wormhole-foundation/wormhole-explorer/common/repository"
	"github.com/wormhole-foundation/wormhole/sdk/vaa"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/bsonrw"
	"go.uber.org/zap"
	"io"
	"log"
	"os"
	"time"
)

func RunBackFillerVaaVolumeV3(vaasBsonFile, outputFile, pricesFile, vaaPayloadParserURL, p2pNetwork string) {

	defer func() {
		fmt.Println("exiting RunBackFillerVaaVolumeV3")
	}()

	loggerInstance := logger.New("wormhole-explorer-analytics", func(cfg *zap.Config) {
		cfg.OutputPaths = []string{"stdout"}
		cfg.ErrorOutputPaths = []string{"stderr"}
	})
	defer loggerInstance.Sync()

	loggerInstance.Info("starting wormhole-explorer-analytics", zap.String("command", "RunBackFillerVaaVolumeV3"))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Open BSON vaasFile
	vaasFile, err := os.Open(vaasBsonFile)
	if err != nil {
		log.Fatal(err)
	}
	defer vaasFile.Close()

	failedVaas, err := os.OpenFile("failed_vaas.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		loggerInstance.Fatal("creating failedVaas file", zap.Error(err))
	}
	defer failedVaas.Close()

	// create a parserVAAAPIClient
	parserVAAAPIClient, err := parser.NewParserVAAAPIClient(10, vaaPayloadParserURL, loggerInstance)
	if err != nil {
		loggerInstance.Fatal("failed to create parse vaa api client")
	}
	tokenResolver := token.NewTokenResolver(parserVAAAPIClient, loggerInstance)
	tokenProvider := domain.NewTokenProvider(p2pNetwork)
	loggerInstance.Info("loading historical prices...")
	priceCache := prices.NewCoinPricesCache(pricesFile)
	priceCache.InitCache()
	converter := NewVaaConverter(priceCache, tokenResolver.GetTransferredTokenByVaa, tokenProvider)

	loggerInstance.Info("loaded historical prices")

	vaasChan := make(chan vaaChanData)
	defer close(vaasChan)

	batchSize := 1
	workersPool := make(chan struct{}, batchSize)
	for i := 0; i < batchSize; i++ {
		workersPool <- struct{}{}
	}

	lpChan := make(chan lpChanData)

	go processVaas(ctx, workersPool, vaasChan, converter, loggerInstance, lpChan, failedVaas)
	go processLineProtocols(ctx, outputFile, loggerInstance, lpChan, failedVaas)

	i := uint64(1)
	offset := uint64(1)

	for {
		// Read the length of the next BSON document (first 4 bytes)
		var docLength int32
		err = binary.Read(vaasFile, binary.LittleEndian, &docLength)
		if err == io.EOF {
			loggerInstance.Info("found end of file,exiting main goroutine...")
			break
		} else if err != nil {
			log.Fatal(err)
		}

		// Read the complete BSON document based on the length
		buffer := make([]byte, docLength)
		binary.LittleEndian.PutUint32(buffer[:4], uint32(docLength)) // Set the length in the buffer
		_, err = io.ReadFull(vaasFile, buffer[4:])
		if err != nil {
			log.Fatal(err)
		}

		if i < offset {
			i++
			continue
		}

		// Create a BSON document reader
		docReader := bsonrw.NewBSONDocumentReader(buffer)
		if err != nil {
			log.Fatal(err)
		}

		// Create a BSON decoder
		decoder, errDecoder := bson.NewDecoder(docReader)
		if errDecoder != nil {
			log.Fatal(errDecoder)
		}

		// Decode the document
		var vaaDoc repository.VaaDoc
		err = decoder.Decode(&vaaDoc)
		if err != nil {
			log.Fatal(err)
		}

		// wait for a worker to be available or exit in case the ctx is cancelled
		select {
		case <-workersPool:
			vaasChan <- vaaChanData{vaa: vaaDoc, i: i}
			break
		case <-ctx.Done():
			return
		}
		i++
	}
	loggerInstance.Info("finished processing vaas")
	loggerInstance.Info("waiting for workers to finish...")
	<-workersPool
	loggerInstance.Info("finished waiting, now cancelling ctx.")
	cancel()

}

func processVaas(ctx context.Context, workersPool chan struct{}, vaasChan chan vaaChanData, converter *VaaConverter, logger *zap.Logger, lpChan chan lpChanData, failedVaas *os.File) {
	for {
		select {
		case vaaData, isOpen := <-vaasChan:
			if !isOpen {
				logger.Info("exiting processVaas,vaasChan is closed")
				return
			}
			go processVaa(ctx, workersPool, vaaData, converter, logger, lpChan, failedVaas)
		case <-ctx.Done():
			logger.Info("exiting processVaas, ctx cancelled")
			return
		}
	}

}

func processVaa(ctx context.Context, workersPool chan struct{}, vaaData vaaChanData, converter *VaaConverter, logger *zap.Logger, lpChan chan lpChanData, failedVaas *os.File) {
	defer func() {
		workersPool <- struct{}{}
	}()
	parsedPayload, point, _, err := converter.Convert(context.Background(), vaaData.vaa.Vaa)

	if err != nil {
		_, errLog := failedVaas.WriteString("vaa: " + vaaData.vaa.ID + "|error: " + err.Error() + "\n")
		if errLog != nil {
			fmt.Printf("\n[processVaa]:failed to write to failedVaas.log. error:" + errLog.Error())
			return
		}
		logger.Error("[processVaa]:failed to convert vaaDoc", zap.Error(err), zap.Uint64("index", vaaData.i))
		return
	}

	m := &metric.Metric{}
	dummyParams := &metric.Params{Vaa: &vaa.VAA{}}
	vaaVolumeV3Point := m.MakePointVaaVolumeV3(point, dummyParams, parsedPayload)
	data := lpChanData{
		vaaId: vaaData.vaa.ID,
		lp:    convertPointToLineProtocol(vaaVolumeV3Point),
		index: vaaData.i,
	}

	select {
	case lpChan <- data:
		break
	case <-ctx.Done():
		logger.Info("exiting processVaa, ctx cancelled")
		break
	}

}

func processLineProtocols(ctx context.Context, outputFile string, logger *zap.Logger, lpChan chan lpChanData, failedVaas *os.File) {

	fout, err := os.OpenFile(outputFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		logger.Fatal("creating output file", zap.Error(err))
	}
	defer fout.Close()

	for {
		select {
		case lpData := <-lpChan:
			_, err = fout.Write([]byte(lpData.lp))
			if err != nil {
				_, errLog := failedVaas.WriteString("vaa: " + lpData.vaaId + "|error: " + err.Error() + "\n")
				if errLog != nil {
					fmt.Printf("\n[%s][processLineProtocols]:failed to write to failedVaas.log. error:%s|output_file_error=%s", time.Now().Format(time.RFC3339), errLog.Error(), err.Error())
					return
				}
				logger.Error("[processLineProtocols]:failed to write line protocol to file", zap.Error(err))
			} else {
				logger.Info("wrote line protocol to file", zap.String("vaaId", lpData.vaaId), zap.Uint64("index", lpData.index))
			}
		case <-ctx.Done():
			logger.Info("exiting processLineProtocols, ctx cancelled")
			return
		}
	}
}

type vaaChanData struct {
	vaa repository.VaaDoc
	i   uint64 // indicate the number of vaa
}

type lpChanData struct {
	vaaId string
	lp    string
	index uint64
}
