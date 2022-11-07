package main

import (
	"context"
	"errors"
	"fly/storage"
	"fmt"
	"os"

	"github.com/certusone/wormhole/node/pkg/common"
	"github.com/certusone/wormhole/node/pkg/p2p"
	gossipv1 "github.com/certusone/wormhole/node/pkg/proto/gossip/v1"
	"github.com/certusone/wormhole/node/pkg/supervisor"
	eth_common "github.com/ethereum/go-ethereum/common"
	crypto2 "github.com/ethereum/go-ethereum/crypto"
	ipfslog "github.com/ipfs/go-log/v2"
	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/wormhole-foundation/wormhole/sdk/vaa"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"

	"github.com/joho/godotenv"
)

var (
	rootCtx       context.Context
	rootCtxCancel context.CancelFunc
)

var (
	p2pNetworkID string
	p2pPort      uint
	p2pBootstrap string
	nodeKeyPath  string
	logLevel     string
)

func main() {
	// Node's main lifecycle context.
	rootCtx, rootCtxCancel = context.WithCancel(context.Background())
	defer rootCtxCancel()
	// main
	p2pNetworkID = "/wormhole/mainnet/2"
	p2pBootstrap = "/dns4/wormhole-mainnet-v2-bootstrap.certus.one/udp/8999/quic/p2p/12D3KooWQp644DK27fd3d4Km3jr7gHiuJJ5ZGmy8hH4py7fP4FP7"
	// devnet
	// p2pNetworkID = "/wormhole/dev"
	// p2pBootstrap = "/dns4/guardian-0.guardian/udp/8999/quic/p2p/12D3KooWL3XJ9EMCyZvmmGXL2LMiVBtrVa2BuESsJiXkSj7333Jw"
	p2pPort = 8999
	nodeKeyPath = "/tmp/node.key"
	logLevel = "warn"
	common.SetRestrictiveUmask()

	lvl, err := ipfslog.LevelFromString(logLevel)
	if err != nil {
		fmt.Println("Invalid log level")
		os.Exit(1)
	}

	logger := ipfslog.Logger("wormhole-fly").Desugar()

	ipfslog.SetAllLoggers(lvl)

	// Verify flags
	if nodeKeyPath == "" {
		logger.Fatal("Please specify --nodeKey")
	}
	if p2pBootstrap == "" {
		logger.Fatal("Please specify --bootstrap")
	}

	// Setup DB
	if err := godotenv.Load(); err != nil {
		logger.Info("No .env file found")
	}
	uri := os.Getenv("MONGODB_URI")
	if uri == "" {
		logger.Fatal("You must set your 'MONGODB_URI' environmental variable. See\n\t https://www.mongodb.com/docs/drivers/go/current/usage-examples/#environment-variable")
	}

	db, err := storage.GetDB(rootCtx, logger, uri, "wormhole")
	if err != nil {
		logger.Fatal("could not connect to DB", zap.Error(err))
	}

	// TODO: change this to use a migration tool.
	isCapped := true
	var sizeCollection, maxDocuments int64 = 500000, 100
	collectionOptions := options.CreateCollectionOptions{
		Capped:       &isCapped,
		SizeInBytes:  &sizeCollection,
		MaxDocuments: &maxDocuments}
	err = db.CreateCollection(context.TODO(), "pyths", &collectionOptions)
	if err != nil {
		target := &mongo.CommandError{}
		isCommandError := errors.As(err, target)
		if !isCommandError || err.(mongo.CommandError).Code != 48 {
			logger.Fatal("error creating pyths capped collection", zap.Error(err))
		}
	}

	repository := storage.NewRepository(db, logger)

	// Outbound gossip message queue
	sendC := make(chan []byte)

	// Inbound observations
	obsvC := make(chan *gossipv1.SignedObservation, 50)

	// Inbound observation requests
	obsvReqC := make(chan *gossipv1.ObservationRequest, 50)

	// Inbound signed VAAs
	signedInC := make(chan *gossipv1.SignedVAAWithQuorum, 50)

	// Heartbeat updates
	heartbeatC := make(chan *gossipv1.Heartbeat, 50)

	// Guardian set state managed by processor
	gst := common.NewGuardianSetState(heartbeatC)

	// Governor cfg
	govConfigC := make(chan *gossipv1.SignedChainGovernorConfig, 50)

	// Governor status
	govStatusC := make(chan *gossipv1.SignedChainGovernorStatus, 50)
	// Bootstrap guardian set, otherwise heartbeats would be skipped
	// TODO: fetch this and probably figure out how to update it live
	gst.Set(&common.GuardianSet{
		Index: 2,
		Keys: []eth_common.Address{
			// mainnet
			eth_common.HexToAddress("0x58CC3AE5C097b213cE3c81979e1B9f9570746AA5"), // Certus One
			eth_common.HexToAddress("0xfF6CB952589BDE862c25Ef4392132fb9D4A42157"), // Staked
			eth_common.HexToAddress("0x114De8460193bdf3A2fCf81f86a09765F4762fD1"), // Figment
			eth_common.HexToAddress("0x107A0086b32d7A0977926A205131d8731D39cbEB"), // ChainodeTech
			eth_common.HexToAddress("0x8C82B2fd82FaeD2711d59AF0F2499D16e726f6b2"), // Inotel
			eth_common.HexToAddress("0x11b39756C042441BE6D8650b69b54EbE715E2343"), // HashQuark
			eth_common.HexToAddress("0x54Ce5B4D348fb74B958e8966e2ec3dBd4958a7cd"), // ChainLayer
			eth_common.HexToAddress("0x66B9590e1c41e0B226937bf9217D1d67Fd4E91F5"), // FTX
			eth_common.HexToAddress("0x74a3bf913953D695260D88BC1aA25A4eeE363ef0"), // Forbole
			eth_common.HexToAddress("0x000aC0076727b35FBea2dAc28fEE5cCB0fEA768e"), // Staking Fund
			eth_common.HexToAddress("0xAF45Ced136b9D9e24903464AE889F5C8a723FC14"), // MoonletWallet
			eth_common.HexToAddress("0xf93124b7c738843CBB89E864c862c38cddCccF95"), // P2P Validator
			eth_common.HexToAddress("0xD2CC37A4dc036a8D232b48f62cDD4731412f4890"), // 01node
			eth_common.HexToAddress("0xDA798F6896A3331F64b48c12D1D57Fd9cbe70811"), // MCF-V2-MAINNET
			eth_common.HexToAddress("0x71AA1BE1D36CaFE3867910F99C09e347899C19C3"), // Everstake
			eth_common.HexToAddress("0x8192b6E7387CCd768277c17DAb1b7a5027c0b3Cf"), // Chorus One
			eth_common.HexToAddress("0x178e21ad2E77AE06711549CFBB1f9c7a9d8096e8"), // syncnode
			eth_common.HexToAddress("0x5E1487F35515d02A92753504a8D75471b9f49EdB"), // Triton
			eth_common.HexToAddress("0x6FbEBc898F403E4773E95feB15E80C9A99c8348d"), // Staking Facilities
			// devnet
			// eth_common.HexToAddress("0xbeFA429d57cD18b7F8A4d91A2da9AB4AF05d0FBe"),
		},
	})

	// Ignore observation requests
	// Note: without this, the whole program hangs on observation requests
	discardMessages(rootCtx, obsvReqC)

	// Log observations
	go func() {
		for {
			select {
			case <-rootCtx.Done():
				return
			case o := <-obsvC:
				ok := verifyObservation(logger, o, gst.Get())
				if !ok {
					logger.Error("Could not verify observation", zap.String("id", o.MessageId))
					continue
				}
				err := repository.UpsertObservation(o)
				if err != nil {
					logger.Error("Error inserting observation", zap.Error(err))
				}
			}
		}
	}()

	// Log signed VAAs
	go func() {
		for {
			select {
			case <-rootCtx.Done():
				return
			case sVaa := <-signedInC:
				v, err := vaa.Unmarshal(sVaa.Vaa)
				if err != nil {
					logger.Error("Error unmarshalling vaa", zap.Error(err))
					continue
				}
				if err := v.Verify(gst.Get().Keys); err != nil {
					logger.Error("Received invalid vaa", zap.String("id", v.MessageID()))
					continue
				}

				if vaa.ChainIDPythNet == v.EmitterChain {
					// handle special logic to Pyth VAA.
					err = repository.UpsertPyth(v, sVaa.Vaa)
					if err != nil {
						logger.Error("Error inserting pyth vaa", zap.Error(err))
					}
				} else {
					// common logic for generic VAA.
					err = repository.UpsertVaa(v, sVaa.Vaa)
					if err != nil {
						logger.Error("Error inserting vaa", zap.Error(err))
					}
				}
			}
		}
	}()

	// Log heartbeats
	go func() {
		for {
			select {
			case <-rootCtx.Done():
				return
			case hb := <-heartbeatC:
				err := repository.UpsertHeartbeat(hb)
				if err != nil {
					logger.Error("Error inserting heartbeat", zap.Error(err))
				}
			}
		}
	}()

	// Log govConfigs
	go func() {
		for {
			select {
			case <-rootCtx.Done():
				return
			case govConfig := <-govConfigC:
				err := repository.UpsertGovernorConfig(govConfig)
				if err != nil {
					logger.Error("Error inserting gov config", zap.Error(err))
				}
			}
		}
	}()

	// Log govStatus
	go func() {
		for {
			select {
			case <-rootCtx.Done():
				return
			case govStatus := <-govStatusC:
				err := repository.UpsertGovernorStatus(govStatus)
				if err != nil {
					logger.Error("Error inserting gov status", zap.Error(err))
				}
			}
		}
	}()

	// Load p2p private key
	var priv crypto.PrivKey
	priv, err = common.GetOrCreateNodeKey(logger, nodeKeyPath)
	if err != nil {
		logger.Fatal("Failed to load node key", zap.Error(err))
	}

	// Run supervisor.
	supervisor.New(rootCtx, logger, func(ctx context.Context) error {
		if err := supervisor.Run(ctx, "p2p", p2p.Run(obsvC, obsvReqC, nil, sendC, signedInC, priv, nil, gst, p2pPort, p2pNetworkID, p2pBootstrap, "", false, rootCtxCancel, nil, govConfigC, govStatusC)); err != nil {
			return err
		}

		logger.Info("Started internal services")

		<-ctx.Done()
		return nil
	},
		// It's safer to crash and restart the process in case we encounter a panic,
		// rather than attempting to reschedule the runnable.
		supervisor.WithPropagatePanic)

	<-rootCtx.Done()
	logger.Info("root context cancelled, exiting...")
	// TODO: wait for things to shut down gracefully
}

func verifyObservation(logger *zap.Logger, obs *gossipv1.SignedObservation, gs *common.GuardianSet) bool {
	pk, err := crypto2.Ecrecover(obs.GetHash(), obs.GetSignature())
	if err != nil {
		return false
	}

	theirAddr := eth_common.BytesToAddress(obs.GetAddr())
	signerAddr := eth_common.BytesToAddress(crypto2.Keccak256(pk[1:])[12:])
	if theirAddr != signerAddr {
		logger.Error("error validating observation, signer addr and addr don't match",
			zap.String("id", obs.MessageId),
			zap.String("obs_addr", theirAddr.Hex()),
			zap.String("signer_addr", signerAddr.Hex()),
		)
		return false
	}

	_, isFromGuardian := gs.KeyIndex(theirAddr)
	if !isFromGuardian {
		logger.Error("error validating observation, signer not in guardian set",
			zap.String("id", obs.MessageId),
			zap.String("obs_addr", theirAddr.Hex()),
		)
	}
	return isFromGuardian
}

func discardMessages[T any](ctx context.Context, obsvReqC chan T) {
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-obsvReqC:
			}
		}
	}()
}
