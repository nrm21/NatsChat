package main

import (
	"context"
	"log"
	"time"

	"github.com/etcd-io/etcd/pkg/transport"
	"go.etcd.io/etcd/clientv3"
)

// ConnToEtcd connects to an ETCD database using TLS settings and returns the
// connection object
func ConnToEtcd(conf etcdConfig) *clientv3.Client {
	tlsInfo := transport.TLSInfo{
		CertFile:      conf.PeerCert,
		KeyFile:       conf.PeerKey,
		TrustedCAFile: conf.CertCa,
	}
	tlsConfig, err := tlsInfo.ClientConfig()
	if err != nil {
		log.Fatal(err)
	}

	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   conf.Endpoints,
		DialTimeout: 5 * time.Second,
		TLS:         tlsConfig,
	})
	if err != nil {
		log.Fatal(err)
	}

	return cli
}

// ReadFromEtcd reads from a key
func ReadFromEtcd(conf etcdConfig, keyToRead string) []string {
	cli := ConnToEtcd(conf)
	defer cli.Close()

	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	gr, err := cli.Get(ctx, keyToRead, clientv3.WithPrefix())
	if err != nil {
		log.Fatal(err)
	}

	var answer []string
	for i := range gr.Kvs {
		answer = append(answer, string(gr.Kvs[i].Value))
	}

	return answer
}

// WriteToEtcd writes to one key in etcd
func WriteToEtcd(conf etcdConfig, keyToWrite string, valueToWrite string) {
	cli := ConnToEtcd(conf)
	defer cli.Close()

	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	_, err := cli.Put(ctx, keyToWrite, valueToWrite)
	if err != nil {
		log.Fatal(err)
	}
}
