package metrics

import (
	"context"
	"fmt"
	"github.com/influxdata/influxdb-client-go/v2/api/write"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	"os"
	"strconv"
	"time"

	"github.com/wormhole-foundation/wormhole-explorer/common/dbutil"
	"github.com/wormhole-foundation/wormhole-explorer/common/logger"
	"go.uber.org/zap"
)

func RunVaaVolumeV3BackFillerFromMongo(mongoUri, mongoDb, outputDir string) {

	ctx := context.Background()
	logInstance := logger.New("wormhole-explorer-analytics")
	logInstance.Info("starting wormhole-explorer-analytics...", zap.String("command", "RunVaaVolumeV3BackFillerFromMongo"))

	db, err := dbutil.Connect(ctx, logInstance, mongoUri, mongoDb, false)
	if err != nil {
		logInstance.Fatal("Failed to connect MongoDB", zap.Error(err))
	}
	defer func() {
		logInstance.Info("closing MongoDB connection...")
		db.DisconnectWithTimeout(10 * time.Second)
	}()

	// Create line protocol file for vaa_volume_v3 in outputDir which is in a persistent-volume
	errLogFileName := fmt.Sprintf("%s/failed_vaas_%s.log", outputDir, time.Now().Format(time.RFC3339))
	errLogFile, err := os.Create(errLogFileName)
	if err != nil {
		logInstance.Fatal("creating failed_vaas.log file", zap.Error(err))
	}
	defer errLogFile.Close()

	batchSize := int32(5000)
	vaasChan := make(chan parsedVaa, batchSize)
	defer close(vaasChan)

	processVaas(vaasChan, outputDir, logInstance)

	filter := bson.M{"rawStandardizedProperties.appIds": bson.M{"$ne": nil}}
	projection := bson.M{"rawStandardizedProperties": 1, "timestamp": 1, "_id": 0}
	findOptions := options.Find().
		SetProjection(projection).
		SetBatchSize(batchSize)
	cur, err := db.Database.Collection("parsedVaa").Find(ctx, filter, findOptions)
	if err != nil {
		logInstance.Fatal("Failed to find VAAs", zap.Error(err))
	}
	defer cur.Close(ctx)

	for cur.Next(ctx) {
		var vaa parsedVaa
		if err = cur.Decode(&vaa); err != nil {
			_, err = errLogFile.WriteString("Failed to decode VAA" + fmt.Sprintf("%v", err) + fmt.Sprintf("%v", cur.Current) + "\n")
			if err != nil {
				logInstance.Fatal("Failed to write failed VAA to log file", zap.Error(err))
			}
			continue
		}
		vaasChan <- vaa
	}

	logInstance.Info("finished wormhole-explorer-analytics", zap.String("command", "RunVaaVolumeV3BackFillerFromMongo"))

}

func processVaas(vaasChan <-chan parsedVaa, outputDir string, logger *zap.Logger) {

	// Create line protocol file for vaa_volume_v3 in outputDir which is in a persistent-volume
	vaaVolumeV3Filename := fmt.Sprintf("%s/vaa_volume_v3_%s.lp", outputDir, time.Now().Format(time.RFC3339))
	vaaVolumeV3File, err := os.Create(vaaVolumeV3Filename)
	if err != nil {
		logger.Fatal("creating vaa_volume_v3 file", zap.Error(err))
	}
	defer vaaVolumeV3File.Close()

	go func() {
		for vaa := range vaasChan {
			point := createInfluxPoint(vaa)
			lp := convertPointToLineProtocol(point)
			_, err = vaaVolumeV3File.WriteString(lp)
			if err != nil {
				logger.Fatal("writing to vaa_volume_v3 file", zap.Error(err))
			}
		}
	}()
}

func createInfluxPoint(vaa parsedVaa) *write.Point {
	point := write.NewPointWithMeasurement("vaa_volume_v3").
		AddTag("version", "v5").
		AddTag("emitter_chain", strconv.Itoa(vaa.RawStandardizedProperties.FromChain)).
		AddTag("destination_chain", strconv.Itoa(vaa.RawStandardizedProperties.ToChain)).
		AddTag("token_chain", strconv.Itoa(vaa.RawStandardizedProperties.TokenChain)).
		AddTag("token_address", vaa.RawStandardizedProperties.TokenAddress).
		AddTag("size", strconv.Itoa(len(vaa.RawStandardizedProperties.AppIDs))).
		AddField("amount", vaa.RawStandardizedProperties.Amount).
		SetTime(vaa.Timestamp)

	for i, appID := range vaa.RawStandardizedProperties.AppIDs {
		point.AddTag(fmt.Sprintf("app_id_%d", i+1), appID)
	}
	for i := len(vaa.RawStandardizedProperties.AppIDs); i < 3; i++ {
		point.AddTag(fmt.Sprintf("app_id_%d", i+1), "none")
	}
	return point
}

type parsedVaa struct {
	RawStandardizedProperties struct {
		AppIDs       []string `bson:"appIds"`
		FromChain    int      `bson:"fromChain"`
		FromAddress  string   `bson:"fromAddress"`
		ToChain      int      `bson:"toChain"`
		TokenChain   int      `bson:"tokenChain"`
		TokenAddress string   `bson:"tokenAddress"`
		Amount       string   `bson:"amount"`
	} `bson:"rawStandardizedProperties"`
	Timestamp time.Time `bson:"timestamp"`
}
