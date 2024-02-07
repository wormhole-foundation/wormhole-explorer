package queue

import gossipv1 "github.com/certusone/wormhole/node/pkg/proto/gossip/v1"

func toObservation(o *gossipv1.SignedObservation) Observation {
	return Observation{
		Addr:      o.Addr,
		Hash:      o.Hash,
		Signature: o.Signature,
		TxHash:    o.TxHash,
		MessageID: o.MessageId,
	}
}

func fromObservation(o *Observation) *gossipv1.SignedObservation {
	return &gossipv1.SignedObservation{
		Addr:      o.Addr,
		Hash:      o.Hash,
		Signature: o.Signature,
		TxHash:    o.TxHash,
		MessageId: o.MessageID,
	}
}
