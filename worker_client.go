package worker


type Client struct {
	producer Producer
}

func (p *Client) Push(jobs ...*Job)error {
	return p.producer.Publish(jobs...)
}

func newClient(pa *Factory) (*Client, error) {
	p, err := pa.producerCreator()
	if err != nil {
		return nil, err
	}
	c := Client{
		producer:p,
	}
	return &c, nil
}
