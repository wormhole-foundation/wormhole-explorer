package gossip

import (
	"github.com/certusone/wormhole/node/pkg/common"
	gossipv1 "github.com/certusone/wormhole/node/pkg/proto/gossip/v1"
)

type GossipChannels struct {

	// Outbound gossip message queue
	SendChannel chan []byte

	// Inbound observations
	ObsvChannel chan *common.MsgWithTimeStamp[gossipv1.SignedObservation]

	// Inbound observation requests - we don't add a environment because we are going to delete this channel
	ObsvReqChannel chan *gossipv1.ObservationRequest

	// Inbound signed VAAs
	SignedInChannel chan *gossipv1.SignedVAAWithQuorum

	// Heartbeat updates
	HeartbeatChannel chan *gossipv1.Heartbeat

	// Governor cfg
	GovConfigChannel chan *gossipv1.SignedChainGovernorConfig

	// Governor status
	GovStatusChannel chan *gossipv1.SignedChainGovernorStatus
}
