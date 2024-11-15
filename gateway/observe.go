package gateway


type RequestResponseDetailsPublisher interface {
	Publish(observe.RequestResponseDetails) error
}
