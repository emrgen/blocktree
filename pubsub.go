package blocktree

type PublishSyncBlocks interface {
	Publish(*SyncBlocks) error
}

type SubscribeSyncBlocks interface {
	Subscribe() <-chan *SyncBlocks
}

type NullPublisher struct {
}

func NewNullPublisher() *NullPublisher {
	return &NullPublisher{}
}

func (n *NullPublisher) Publish(*SyncBlocks) error {
	return nil
}
