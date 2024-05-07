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
	groupID := createDeduplicationIDForDuplicateVaa(e)
	_, err = s.api.Publish(ctx,
		&aws_sns.PublishInput{
			MessageGroupId:         aws.String(groupID),
			MessageDeduplicationId: aws.String(groupID),
			Message:                aws.String(string(body)),
			TopicArn:               aws.String(s.url),
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
