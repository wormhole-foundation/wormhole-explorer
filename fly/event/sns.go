package event

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	aws_sns "github.com/aws/aws-sdk-go-v2/service/sns"
	"github.com/wormhole-foundation/wormhole-explorer/fly/internal/track"
)

type SnsEventDispatcher struct {
	api *aws_sns.Client
	url string
}

func NewSnsEventDispatcher(awsConfig aws.Config, url string) (*SnsEventDispatcher, error) {
	return &SnsEventDispatcher{
		api: aws_sns.NewFromConfig(awsConfig),
		url: url,
	}, nil
}

func (s *SnsEventDispatcher) NewDuplicateVaa(ctx context.Context, e DuplicateVaa) error {
	body, err := json.Marshal(event{
		TrackID: track.GetTrackIDForDuplicatedVAA(e.VaaID),
		Type:    "duplicated-vaa",
		Source:  "fly",
		Data:    e,
	})
	if err != nil {
		return err
	}
	groupID := fmt.Sprintf("%s-%s", e.VaaID, e.Digest)
	_, err = s.api.Publish(ctx,
		&aws_sns.PublishInput{
			MessageGroupId:         aws.String(groupID),
			MessageDeduplicationId: aws.String(groupID),
			Message:                aws.String(string(body)),
			TopicArn:               aws.String(s.url),
		})
	return err
}

func (s *SnsEventDispatcher) NewGovernorStatus(ctx context.Context, e GovernorStatus) error {
	body, err := json.Marshal(event{
		TrackID: track.GetTrackIDForGovernorStatus(e.NodeName, e.Timestamp),
		Type:    "governor-status",
		Source:  "fly",
		Data:    e,
	})
	if err != nil {
		return err
	}
	grouID := fmt.Sprintf("%s-%v", e.NodeAddress, e.Timestamp)
	_, err = s.api.Publish(ctx,
		&aws_sns.PublishInput{
			MessageGroupId:         aws.String(grouID),
			MessageDeduplicationId: aws.String(grouID),
			Message:                aws.String(string(body)),
			TopicArn:               aws.String(s.url),
		})
	return err
}
