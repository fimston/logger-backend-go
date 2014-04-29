package destination

import (
	"github.com/dmotylev/goproperties"
)

type Destination interface {
	Push([]byte) error
}

type Destinations struct {
	dests []Destination
}

func NewDestinations(config *properties.Properties) (*Destinations, error) {
	rabbit, err := NewRabbitMq(config)
	if err != nil {
		return nil, err
	}

	dest := &Destinations{}
	dest.register(rabbit)
	return dest, nil
}

func (self *Destinations) register(dest Destination) {
	self.dests = append(self.dests, dest)
}

func (self *Destinations) Push(d []byte) error {
	for _, dst := range self.dests {
		err := dst.Push(d)
		if err != nil {
			return err
		}
	}
	return nil
}
