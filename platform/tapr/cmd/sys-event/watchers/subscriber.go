package watchers

type Subscriber struct {
	Watchers     *Watchers
	Notification *Notification
}

func NewSubscriber(watchers *Watchers) *Subscriber {
	return &Subscriber{Watchers: watchers}
}

func (s *Subscriber) WithNotification(n *Notification) *Subscriber {
	s.Notification = n
	return s
}
