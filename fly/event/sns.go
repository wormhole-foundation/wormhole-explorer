package event

import (
	"context"
	"crypto/sha512"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"

	"github.com/aws/aws-sdk-go-v2/aws"
	aws_sns "github.com/aws/aws-sdk-go-v2/service/sns"
	"github.com/aws/aws-sdk-go-v2/service/sns/types"
	"github.com/wormhole-foundation/wormhole-explorer/fly/internal/track"
)

type SnsEventDispatcher struct {
	api          *aws_sns.Client
	processorUrl string
	pipelineUrl  string
}

func NewSnsEventDispatcher(awsConfig aws.Config, processorUrl string, pipelineUrl string) (*SnsEventDispatcher, error) {
	return &SnsEventDispatcher{
		api:          aws_sns.NewFromConfig(awsConfig),
		processorUrl: processorUrl,
		pipelineUrl:  pipelineUrl,
	}, nil
}

func (s *SnsEventDispatcher) NewDuplicateVaa(ctx context.Context, e DuplicateVaa) error {
	attrs := map[string]types.MessageAttributeValue{
		"messageType": {
			DataType:    aws.String("String"),
			StringValue: aws.String("duplicated-vaa"),
		},
	}
	body, err := json.Marshal(event{
		TrackID: track.GetTrackIDForDuplicatedVAA(e.VaaID),
		Type:    "duplicated-vaa",
		Source:  "fly",
		Data:    e,
	})
	if err != nil {
		return err
	}
	groupID := createDeduplicationIDForDuplicateVaa(e)
	_, err = s.api.Publish(ctx,
		&aws_sns.PublishInput{
			MessageGroupId:         aws.String(groupID),
			MessageDeduplicationId: aws.String(groupID),
			Message:                aws.String(string(body)),
			TopicArn:               aws.String(s.processorUrl),
			MessageAttributes:      attrs,
		})
	return err
}

func createDeduplicationIDForDuplicateVaa(e DuplicateVaa) string {
	id := fmt.Sprintf("%s%s", e.Digest, e.VaaID)
	h := sha512.New()
	io.WriteString(h, id)
	deduplicationID := base64.StdEncoding.EncodeToString(h.Sum(nil))
	if len(deduplicationID) > 127 {
		return deduplicationID[:127]
	}
	return deduplicationID
}

func (s *SnsEventDispatcher) NewGovernorStatus(ctx context.Context, e GovernorStatus) error {
	attrs := map[string]types.MessageAttributeValue{
		"messageType": {
			DataType:    aws.String("String"),
			StringValue: aws.String("governor"),
		},
	}
	body, err := json.Marshal(event{
		TrackID: track.GetTrackIDForGovernorStatus(e.NodeName, e.Timestamp),
		Type:    "governor-status",
		Source:  "fly",
		Data:    e,
	})
	if err != nil {
		return err
	}
	groupID := "governor-status"
	deduplicationID := fmt.Sprintf("gov-status-%s-%v", e.NodeAddress, e.Timestamp)
	_, err = s.api.Publish(ctx,
		&aws_sns.PublishInput{
			MessageGroupId:         aws.String(groupID),
			MessageDeduplicationId: aws.String(deduplicationID),
			Message:                aws.String(string(body)),
			TopicArn:               aws.String(s.processorUrl),
			MessageAttributes:      attrs,
		})
	return err
}

func (s *SnsEventDispatcher) NewGovernorConfig(ctx context.Context, e GovernorConfig) error {
	attrs := map[string]types.MessageAttributeValue{
		"messageType": {
			DataType:    aws.String("String"),
			StringValue: aws.String("governor"),
		},
	}
	body, err := json.Marshal(event{
		TrackID: track.GetTrackIDForGovernorConfig(e.NodeName, e.Timestamp),
		Type:    "governor-config",
		Source:  "fly",
		Data:    e,
	})
	if err != nil {
		return err
	}
	groupID := fmt.Sprintf("gov-config-%s-%v", e.NodeAddress, e.Timestamp)
	_, err = s.api.Publish(ctx,
		&aws_sns.PublishInput{
			MessageGroupId:         aws.String(groupID),
			MessageDeduplicationId: aws.String(groupID),
			Message:                aws.String(string(body)),
			TopicArn:               aws.String(s.processorUrl),
			MessageAttributes:      attrs,
		})
	return err
}

func (s *SnsEventDispatcher) NewAttestationVaa(ctx context.Context, vaa Vaa) error {
	attrs := map[string]types.MessageAttributeValue{
		"messageType": {
			DataType:    aws.String("String"),
			StringValue: aws.String("vaa"),
		},
	}
	body, err := json.Marshal(event{
		TrackID: track.GetTrackID(vaa.VaaID),
		Type:    "vaa",
		Source:  "fly",
		Data:    vaa,
	})
	if err != nil {
		return err
	}
	groupID := createDeduplicationIDForVaa(vaa)
	_, err = s.api.Publish(ctx,
		&aws_sns.PublishInput{
			MessageGroupId:         aws.String(groupID),
			MessageDeduplicationId: aws.String(groupID),
			Message:                aws.String(string(body)),
			TopicArn:               aws.String(s.pipelineUrl),
			MessageAttributes:      attrs,
		})
	return err
}

func createDeduplicationIDForVaa(vaa Vaa) string {
	// id = {digest + vaaID}
	id := fmt.Sprintf("%s%s", vaa.ID, vaa.VaaID)
	h := sha512.New()
	io.WriteString(h, id)
	deduplicationID := base64.StdEncoding.EncodeToString(h.Sum(nil))
	if len(deduplicationID) > 127 {
		return deduplicationID[:127]
	}
	return deduplicationID
}
