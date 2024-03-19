package pool

import (
	"sort"
	"time"

	"golang.org/x/net/context"
	"golang.org/x/time/rate"
)

// Pool is a pool of items.
type Pool struct {
	items []Item
}

// Item defines the item of the pool.
type Item struct {
	// Id is the item ID.
	Id string
	// description of the item.
	Description string
	// priority is the priority of the item.
	// The lower the value, the higher the priority.
	priority uint8
	// rateLimit is the rate limiter for the item.
	rateLimit *rate.Limiter
}

// NewPool creates a new pool.
func NewPool(cfg []Config) *Pool {
	p := &Pool{}
	for _, c := range cfg {
		p.addItem(c)
	}
	return p
}

// addItem adds a new item to the pool.
func (p *Pool) addItem(cfg Config) {
	i := Item{
		Id:          cfg.Id,
		Description: cfg.Description,
		priority:    cfg.Priority,
		rateLimit: rate.NewLimiter(
			rate.Every(time.Minute/time.Duration(cfg.RequestsPerMinute)), 1),
	}
	p.items = append(p.items, i)
}

// GetItem returns the next available item of the pool.
func (p *Pool) GetItem() Item {
	// check if there is no item
	if len(p.items) == 0 {
		return Item{}
	}

	// get the next available item
	itemWithScore := []struct {
		item  Item
		score float64
	}{}

	now := time.Now()
	for _, i := range p.items {
		tokenAt := i.rateLimit.TokensAt(now)
		itemWithScore = append(itemWithScore, struct {
			item  Item
			score float64
		}{
			item:  i,
			score: tokenAt,
		})
	}

	// sort by score and priority
	sort.Slice(itemWithScore, func(i, j int) bool {
		if itemWithScore[i].score == itemWithScore[j].score {
			return itemWithScore[i].item.priority < itemWithScore[j].item.priority
		}
		return itemWithScore[i].score > itemWithScore[j].score
	})

	return itemWithScore[0].item
}

// GetItems returns the list of items sorted by score and priority.
// Once there is an event on the item, it must be notified using the method NotifyEvent.
func (p *Pool) GetItems() []Item {
	if len(p.items) == 0 {
		return []Item{}
	}

	itemsWithScore := []struct {
		item  Item
		score float64
	}{}

	now := time.Now()
	for _, i := range p.items {
		tokenAt := i.rateLimit.TokensAt(now)
		itemsWithScore = append(itemsWithScore, struct {
			item  Item
			score float64
		}{
			item:  i,
			score: tokenAt,
		})
	}

	// sort by score and priority
	sort.Slice(itemsWithScore, func(i, j int) bool {
		if itemsWithScore[i].score == itemsWithScore[j].score {
			return itemsWithScore[i].item.priority < itemsWithScore[j].item.priority
		}
		return itemsWithScore[i].score > itemsWithScore[j].score
	})

	// convert itemsWithScore to items
	items := []Item{}
	for _, i := range itemsWithScore {
		items = append(items, i.item)
	}
	return items
}

// Wait waits for the rate limiter to allow the next item request.
func (i *Item) Wait(ctx context.Context) error {
	return i.rateLimit.Wait(ctx)
}
