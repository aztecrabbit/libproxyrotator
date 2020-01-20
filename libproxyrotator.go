package libproxyrotator

import (
	"os"
    "io/ioutil"
    "log"
	"fmt"
	"net"
	"context"

	"golang.org/x/net/proxy"
	"github.com/armon/go-socks5"
	"github.com/aztecrabbit/liblog"
)

type ProxyRotator struct {
	Port string
	Proxies []string
}

func (p *ProxyRotator) RotateProxies() {
	p.Proxies = append(p.Proxies[1:], p.Proxies[0])
}

func (p *ProxyRotator) Start() {
	config := &socks5.Config{
		Logger: log.New(os.Stdout, "", log.LstdFlags),
		Dial: func(ctx context.Context, net_, addr string) (net.Conn, error) {
			var netConn net.Conn
			var lastError error

			remoteProxies := p.Proxies

			p.RotateProxies()

			for _, remoteProxy := range remoteProxies {
				dialer, err := proxy.SOCKS5("tcp", remoteProxy, nil, proxy.Direct)
				if err != nil {
					panic(err)
				}

				data, err := dialer.Dial(net_, addr)
				if err != nil {
					lastError = err
					continue
				}

				return data, err
			}

			return netConn, lastError
		},
	}

	config.Logger.SetOutput(ioutil.Discard)

	server, err := socks5.New(config)
	if err != nil {
		panic(err)
	}

	if err := server.ListenAndServe("tcp", "0.0.0.0:" + p.Port); err != nil {
		liblog.LogInfo(fmt.Sprintf("Exception\n\n|   %v\n|\n", err), "INFO", liblog.Colors["R1"])
		os.Exit(0)
	}
}
