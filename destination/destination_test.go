package destination

import (
	"flag"
	"github.com/dmotylev/goproperties"
	"testing"
	"time"
)

var (
	configPath = flag.String("config", "", "path to configuration file")
)

func TestDestinations(test *testing.T) {

	flag.Parse()

	props, err := properties.Load(*configPath)

	if err != nil {
		test.Fatal(err)
	}

	dest := NewDestinations()

	rab, err := NewRabbitMq(&props)
	if err != nil {
		test.Fatal(err)
	}
	dest.Register(rab)

	err = dest.Push([]byte("hello, world2"))
	if err != nil {
		test.Fatal(err)
	}

	time.Sleep(time.Minute)
}
