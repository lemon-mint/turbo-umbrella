package turboumbrella

import (
	"log"
	"net"
	"os"
	"sync"
	"time"

	virtualnotify "github.com/lemon-mint/VirtualNotify"
	"github.com/libp2p/go-reuseport"
)

type Turboumbrella struct {
	NameSpace string
	AppName   string
	Version   string

	network string
	host    string

	ln net.Listener

	// Callbacks
	OnUpgrade func()

	// PID
	myPid int

	// EventListeners
	vnevs *virtualnotify.VirtualNotify

	// Shutdown Onces
	shutdownOnce sync.Once
}

const (
	upgrade_evt = "upgradeStart"
)

func New(nameSpace, network, host string) (*Turboumbrella, error) {
	ID := "turboumbrella_" + nameSpace
	tu := &Turboumbrella{
		NameSpace: nameSpace,

		network: network,
		host:    host,
	}
	var err error
	tu.ln, err = reuseport.Listen(network, host)
	if err != nil {
		return nil, err
	}

	tu.shutdownOnce = sync.Once{}

	tu.myPid = os.Getpid()

	tu.vnevs = virtualnotify.New(ID)

	return tu, nil
}

func (tu *Turboumbrella) Listener() net.Listener {
	return tu.ln
}

func (tu *Turboumbrella) Close() error {
	var err error
	tu.shutdownOnce.Do(
		func() {
			tu.vnevs.Close()
			err = tu.ln.Close()
		},
	)
	return err
}

func (tu *Turboumbrella) WaitForUpgrade() error {
	err := tu.vnevs.Subscribe(upgrade_evt)
	if err != nil {
		return err
	}
	go tu.vnevs.Run(time.Second)
	log.Println("Waiting for upgrade event")
	for {
		ev, err := tu.vnevs.Next()
		if err != nil {
			return err
		}
		if ev.Name == upgrade_evt {
			log.Println("upgrade event received")

			if tu.OnUpgrade != nil {
				tu.OnUpgrade()
			}

			err = tu.Close()
			if err != nil && err != os.ErrClosed {
				return err
			}

			return nil
		}
	}
}

func (tu *Turboumbrella) Upgrade(timeout time.Duration) error {
	log.Println("Sending upgrade event")
	err := tu.vnevs.PublishTimeout(upgrade_evt, timeout)
	return err
}
