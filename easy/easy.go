package easy

import (
	"flag"
	"fmt"
	"net/http"

	"github.com/p4gefau1t/trojan-go/common"
	"github.com/p4gefau1t/trojan-go/conf"
	"github.com/p4gefau1t/trojan-go/log"
	"github.com/p4gefau1t/trojan-go/proxy"
	"v2ray.com/core/common/net"
)

type EasyOption struct {
	common.OptionHandler

	server   *bool
	client   *bool
	password *string
	local    *string
	remote   *string
	cert     *string
	key      *string
}

func (o *EasyOption) Name() string {
	return "easy"
}

func (o *EasyOption) Handle() error {
	if !*o.server && !*o.client {
		return common.NewError("empty")
	}
	if *o.password == "" {
		log.Fatal("empty password is not allowed")
	}
	if *o.client {
		clientConfigFormat := `
{
    "run_type": "client",
    "local_addr": "%s",
    "local_port": %s,
    "remote_addr": "%s",
    "remote_port": %s,
    "password": [
        "%s"
    ]
}
		`
		if *o.local == "" {
			log.Warn("client local addr is unspecified, using 127.0.0.1:1080")
			*o.local = "127.0.0.1:1080"
		}
		localHost, localPort, err := net.SplitHostPort(*o.local)
		if err != nil {
			log.Fatal(common.NewError("invalid local addr format").Base(err))
		}
		remoteHost, remotePort, err := net.SplitHostPort(*o.remote)
		if err != nil {
			log.Fatal(common.NewError("invalid remote addr format").Base(err))
		}
		clientConfigJSON := fmt.Sprintf(clientConfigFormat, localHost, localPort, remoteHost, remotePort, *o.password)
		log.Debug(clientConfigJSON)
		config, err := conf.ParseJSON([]byte(clientConfigJSON))
		if err != nil {
			log.Fatal(config)
		}
		client, err := proxy.NewProxy(config)
		if err != nil {
			log.Fatal(config)
		}
		err = client.Run()
		if err != nil {
			log.Fatal(err)
		}
	} else if *o.server {
		serverConfigFormat := `
{
    "run_type": "server",
    "local_addr": "%s",
    "local_port": %s,
    "remote_addr": "%s",
    "remote_port": %s,
    "password": [
        "%s"
    ],
    "ssl": {
        "cert": "%s",
        "key": "%s"
    }
}
		`
		if *o.remote == "" {
			log.Fatal("-remote is empty, you should fill in a valid web server address, e.g. 127.0.0.1:80")
		}
		//check web server
		resp, err := http.Get("http://" + *o.remote)
		if err != nil {
			log.Fatal(common.NewError(*o.remote + " is not a valid web server").Base(err))
		}
		buf := [128]byte{}
		_, err = resp.Body.Read(buf[:])
		if err != nil {
			log.Fatal(common.NewError(*o.remote + " is not a valid web server").Base(err))
		}
		log.Debug(string(buf[:]))
		if *o.local == "" {
			log.Warn("server local addr is unspecified, using 0.0.0.0:443")
			*o.local = "0.0.0.0:443"
		}
		localHost, localPort, err := net.SplitHostPort(*o.local)
		if err != nil {
			log.Fatal(common.NewError("invalid local addr format").Base(err))
		}
		remoteHost, remotePort, err := net.SplitHostPort(*o.remote)
		if err != nil {
			log.Fatal(common.NewError("invalid remote addr format").Base(err))
		}
		serverConfigJSON := fmt.Sprintf(serverConfigFormat, localHost, localPort, remoteHost, remotePort, *o.password, *o.cert, *o.key)
		log.Debug(serverConfigJSON)
		config, err := conf.ParseJSON([]byte(serverConfigJSON))
		if err != nil {
			log.Fatal(err)
		}
		server, err := proxy.NewProxy(config)
		if err != nil {
			log.Fatal(err)
		}
		err = server.Run()
		if err != nil {
			log.Fatal(err)
		}
	}
	return nil
}

func (o *EasyOption) Priority() int {
	return 50
}

func init() {
	common.RegisterOptionHandler(&EasyOption{
		server:   flag.Bool("server", false, "Run a trojan-go server"),
		client:   flag.Bool("client", false, "Run a trojan-go client"),
		password: flag.String("password", "", "Password for authentication"),
		remote:   flag.String("remote", "", "Remote address, for example 127.0.0.1:12345"),
		local:    flag.String("local", "", "Local address, for example 127.0.0.1:12345"),
		key:      flag.String("key", "server.key", "Key of the server"),
		cert:     flag.String("cert", "server.crt", "Certificates of the server"),
	})
}
