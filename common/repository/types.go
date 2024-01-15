package repository

import (
	"time"
)

type IndexingTimestamps struct {
	IndexedAt time.Time `bson:"indexedAt"`
}

func IndexedAt(t time.Time) IndexingTimestamps {
	return IndexingTimestamps{
		IndexedAt: t,
	}
}
