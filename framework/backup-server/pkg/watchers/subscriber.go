package watchers

type Subscriber struct {
	Watchers *Watchers
}

func NewSubscriber(watchers *Watchers) *Subscriber {
	return &Subscriber{Watchers: watchers}
}
