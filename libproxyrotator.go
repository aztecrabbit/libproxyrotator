package libproxyrotator

import (
	"context"
	"errors"
	"io/ioutil"
	"log"
	"net"
	"os"

	"github.com/armon/go-socks5"
	"github.com/aztecrabbit/liblog"
	"github.com/aztecrabbit/libutils"
	"golang.org/x/net/proxy"
)

var (
	DefaultConfig = &Config{
		Port: "3080",
	}
)

type Config struct {
	Port string
}

type ProxyRotator struct {
	Config  *Config
	Proxies []string
}

func (p *ProxyRotator) AddProxy(address string) {
	p.Proxies = append(p.Proxies, address)
}

func (p *ProxyRotator) GetProxy() (string, error) {
	libutils.Lock.Lock()
	defer libutils.Lock.Unlock()

	if len(p.Proxies) == 0 {
		return "", errors.New("Proxy Address not found")
	}

	proxyAddress := p.Proxies[0]

	if len(p.Proxies) > 1 {
		p.Proxies = append(p.Proxies[1:], p.Proxies[0])
	}

	return proxyAddress, nil
}

func (p *ProxyRotator) DeleteProxy(address string) {
	for i, proxyAddress := range p.Proxies {
		if proxyAddress == address {
			p.Proxies = append(p.Proxies[:i], p.Proxies[i+1:]...)
			break
		}
	}
}

func (p *ProxyRotator) Start() {
	config := &socks5.Config{
		Logger: log.New(os.Stdout, "", log.LstdFlags),
		Dial: func(ctx context.Context, net_, addr string) (net.Conn, error) {
			for i := 0; i < len(p.Proxies); i++ {
				proxyAddress, err := p.GetProxy()
				if err != nil {
					break
				}

				dialer, err := proxy.SOCKS5("tcp", proxyAddress, nil, proxy.Direct)
				if err != nil {
					panic(err)
				}

				data, err := dialer.Dial(net_, addr)
				if err != nil {
					continue
				}

				return data, nil
			}

			return nil, errors.New("proxies not available")
		},
	}

	config.Logger.SetOutput(ioutil.Discard)

	server, err := socks5.New(config)
	if err != nil {
		panic(err)
	}

	if err := server.ListenAndServe("tcp", "0.0.0.0:"+p.Config.Port); err != nil {
		liblog.LogException(err, "INFO")
		os.Exit(0)
	}
}
