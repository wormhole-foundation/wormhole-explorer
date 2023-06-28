package notifier

import (
	"context"
	"fmt"
	"strconv"

	"github.com/go-redis/redis/v8"
	"github.com/wormhole-foundation/wormhole/sdk/vaa"
)

const LUA_SCRIPT = `
local newValue = ARGV[1];
if (newValue == "" or newValue:find("%D")) then
	return redis.error_reply(string.format("[%s] is not a valid number", newValue));
end
local currentValue = redis.call('get', KEYS[1]);
if currentValue then
	if string.len(newValue) > string.len(currentValue) then
		redis.call('set', KEYS[1], ARGV[1]);
		return newValue
	elseif string.len(newValue) < string.len(currentValue) then
		return currentValue;
	elseif newValue > currentValue then
		redis.call('set', KEYS[1], ARGV[1])
		return newValue
	else
		return currentValue
	end
else
	redis.call('set', KEYS[1], ARGV[1])
	return newValue
end
`

type LastSequenceNotifier struct {
	client *redis.Client
	script *redis.Script
	prefix string
}

func NewLastSequenceNotifier(c *redis.Client, prefix string) *LastSequenceNotifier {
	if prefix == "" {
		prefix = "wormscan:vaa-max-sequence"
	} else {
		prefix = fmt.Sprintf("%s-wormscan:vaa-max-sequence", prefix)
	}

	return &LastSequenceNotifier{
		client: c,
		script: redis.NewScript(LUA_SCRIPT),
		prefix: prefix,
	}
}

func (l *LastSequenceNotifier) Notify(ctx context.Context, v *vaa.VAA, _ []byte) error {
	key := fmt.Sprintf("%s:%d:%s", l.prefix, v.EmitterChain, v.EmitterAddress.String())
	sequence := strconv.FormatUint(v.Sequence, 10)
	_, err := l.script.Run(ctx, l.client, []string{key}, sequence).Result()
	return err
}
