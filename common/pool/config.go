package pool

// Config is the configuration of an pool item.
type Config struct {
	// id is the RPC service ID.
	Id string
	// priority is the priority of the item.
	Priority uint8
	// amount of request per minute
	RequestsPerMinute uint
}
