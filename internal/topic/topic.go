package topic

type Topic struct {
	TopicID uint16
	MQ      Queue
	Cgroups []CGroup
}

func (t *Topic) Init(tid uint16) {
	t.TopicID = tid
	t.MQ.Init()
	t.Cgroups = make([]CGroup, 0)
}
