package topic

type Topic struct {
	TopicID uint16
	MQ      Queue
}

func (t *Topic) Init(tid uint16) {
	t.TopicID = tid
	t.MQ.Init()
}
