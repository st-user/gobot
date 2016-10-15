package megapi

import (
	"io"
	"time"

	"github.com/hybridgroup/gobot"
	"github.com/tarm/serial"
)

var _ gobot.Adaptor = (*Adaptor)(nil)

// Adaptor is the Gobot adaptor for the MakeBlock MegaPi board
type Adaptor struct {
	name              string
	connection        io.ReadWriteCloser
	serialConfig      *serial.Config
	writeBytesChannel chan []byte
	finalizeChannel   chan struct{}
}

// NewAdaptor returns a new MegaPi Adaptor with specified serial port used to talk to the MegaPi with a baud rate of 115200
func NewAdaptor(device string) *Adaptor {
	c := &serial.Config{Name: device, Baud: 115200}
	return &Adaptor{
		name:              "MegaPi",
		connection:        nil,
		serialConfig:      c,
		writeBytesChannel: make(chan []byte),
		finalizeChannel:   make(chan struct{}),
	}
}

// Name returns the name of this adaptor
func (megaPi *Adaptor) Name() string {
	return megaPi.name
}

// SetName sets the name of this adaptor
func (megaPi *Adaptor) SetName(n string) {
	megaPi.name = n
}

// Connect starts a connection to the board
func (megaPi *Adaptor) Connect() (errs []error) {
	if megaPi.connection == nil {
		sp, err := serial.OpenPort(megaPi.serialConfig)
		if err != nil {
			return []error{err}
		}

		// sleeping is required to give the board a chance to reset
		time.Sleep(2 * time.Second)
		megaPi.connection = sp
	}

	// kick off thread to send bytes to the board
	go func() {
		for {
			select {
			case bytes := <-megaPi.writeBytesChannel:
				megaPi.connection.Write(bytes)
				time.Sleep(10 * time.Millisecond)
			case <-megaPi.finalizeChannel:
				megaPi.finalizeChannel <- struct{}{}
				return
			default:
				time.Sleep(10 * time.Millisecond)
			}
		}
	}()
	return
}

// Finalize terminates the connection to the board
func (megaPi *Adaptor) Finalize() (errs []error) {
	megaPi.finalizeChannel <- struct{}{}
	<-megaPi.finalizeChannel
	if err := megaPi.connection.Close(); err != nil {
		return []error{err}
	}
	return
}