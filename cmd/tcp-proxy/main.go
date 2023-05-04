package main

import (
	"flag"
	proxy "github.com/jpillora/go-tcp-proxy"
	"net"
	"os"
)

var (
	version = "1.1.0-src"

	localAddr   = flag.String("l", ":9999", "local address")
	remoteAddr  = flag.String("r", "localhost:80", "remote address")
	verbose     = flag.Bool("v", false, "display server actions")
	veryverbose = flag.Bool("vv", false, "display server actions and all tcp data")
	nagles      = flag.Bool("n", false, "disable nagles algorithm")
	hex         = flag.Bool("h", false, "output hex")
	colors      = flag.Bool("c", false, "output ansi colors")
	unwrapTLS   = flag.Bool("unwrap-tls", false, "remote connection with TLS exposed unencrypted locally")
	tlsEnabled  = flag.Bool("tls", false, "tls enabled")
	match       = flag.String("match", "", "match regex (in the form 'regex')")
	replace     = flag.String("replace", "", "replace regex (in the form 'regex~replacer')")
	rootCert    = flag.String("cert", "", "location of pem certificate")
)

func main() {
	flag.Parse()

	logger := proxy.ColorLogger{
		Verbose: *verbose,
		Color:   *colors,
	}

	logger.Info("go-tcp-proxy (%s) proxing from %v to %v ", version, *localAddr, *remoteAddr)

	laddr, err := net.ResolveTCPAddr("tcp", *localAddr)
	if err != nil {
		logger.Warn("Failed to resolve local address: %s", err)
		os.Exit(1)
	}
	raddr, err := net.ResolveTCPAddr("tcp", *remoteAddr)
	if err != nil {
		logger.Warn("Failed to resolve remote address: %s", err)
		os.Exit(1)
	}
	listener, err := net.ListenTCP("tcp", laddr)
	if err != nil {
		logger.Warn("Failed to open local port to listen: %s", err)
		os.Exit(1)
	}

	var certloc string
	if *tlsEnabled {
		if *rootCert == "" {
			logger.Warn("TLS enabled but no certificate provided, using default rds certs")
			certloc = "/tmp/rds-combined-ca-bundle.pem"
		} else {
			certloc = *rootCert
		}
	}

	tcpProxy := &proxy.TcpProxy{
		Logger:           logger,
		Listener:         listener,
		LocalAddrString:  *localAddr,
		RemoteAddrString: *remoteAddr,
		Match:            *match,
		LocalAddr:        laddr,
		RemoteAddr:       raddr,
		Verbose:          *veryverbose,
		Replace:          *replace,
		UnWrapTLS:        *unwrapTLS,
		Nagle:            *nagles,
		Hex:              *hex,
		Color:            *colors,
		RootCertLocation: certloc,
	}

	if err := tcpProxy.Start(); err != nil {
		logger.Warn("Failed to start proxy: %s", err)
		os.Exit(1)
	}
}
