package main

import (
	"context"
	"flag"
	"github.com/jxskiss/gothrifter/example/search/gen-thrifter/search"
	"github.com/jxskiss/gothrifter/lib/go-kit"
	"github.com/jxskiss/gothrifter/lib/thrift"
	"log"
	"time"
)

//go:generate go run ../../main.go -prefix github.com/jxskiss/gothrifter/example/search -idl service.thrift

func main() {
	server := flag.Bool("server", false, "run server")
	usekit := flag.Bool("kit", false, "using go-kit")
	addr := flag.String("addr", "127.0.0.1:9090", "address")
	flag.Parse()

	var err error
	if *server {
		if *usekit {
			err = runKitServer(*addr)
		} else {
			err = runSimpleServer(*addr)
		}
		if err != nil {
			log.Println("server error:", err)
		}
	} else {
		if *usekit {
			err = runKitClient()
		} else {
			err = runSimpleClient(*addr)
		}
		if err != nil {
			log.Println("client error:", err)
		}
	}
}

func runSimpleServer(addr string) error {
	processor := search.NewSearchServiceProcessor(&searchImpl{})
	server := thrift.NewServer(processor, thrift.WithCompact())
	return server.ListenAndServe(addr)
}

func runSimpleClient(addr string) error {
	// We can use binary protocol to talk to server using compact protocol as default,
	// the server will automatically recognize which protocol client is using.
	tclient := thrift.NewClient(thrift.StdDialer, addr)
	cli := search.NewSearchServiceClient(tclient)
	return doRequests(cli)
}

func doRequests(cli search.SearchServiceHandler) error {
	req := &search.SearchRequest{
		Query:         "dummy query",
		PageNumber:    2,
		ResultPerPage: 20,
	}
	for i := 0; i < 5; i++ {
		resp, err := cli.Search(context.Background(), req)
		if err != nil {
			return err
		}
		log.Println("search response=", resp, "err=", err)

		err = cli.Ping(context.Background())
		log.Println("ping err=", err)

		err = cli.Ack(context.Background(), 1234)
		log.Println("ack err=", err)
	}

	return nil
}

// go-kit server and client

func runKitServer(addr string) error {
	registrar, err := kit.SimpleConsulRegistrar("svc_search_instance0", "svc_search", addr)
	if err != nil {
		return err
	}

	processor := search.NewSearchServiceKitDefaultProcessor("svc_search", &searchImpl{})
	server := thrift.NewServer(processor, thrift.WithHeader())

	if err := server.Listen(addr); err != nil {
		return err
	}
	registrar.Register()
	defer registrar.Deregister()
	return server.Serve()
}

func runKitClient() error {
	cli, err := search.NewSearchServiceKitClientSimpleConsul("dummy_client", "svc_search")
	if err != nil {
		return err
	}
	return doRequests(cli)
}

// Service implementation.

type searchImpl struct {
}

func (*searchImpl) Search(ctx context.Context, req *search.SearchRequest) (*search.SearchResponse, error) {
	log.Println("*searchImpl.Search called")
	resp := &search.SearchResponse{
		Results: []*search.Result{
			{
				Url:      "dummy url",
				Title:    "dummy title",
				Snippets: []string{"dummy", "snippets"},
				PostAt:   time.Now().Unix(),
			},
		},
	}
	return resp, nil
}

func (*searchImpl) Ping(ctx context.Context) error {
	log.Println("*searchImpl.Ping called")
	return nil
}

func (*searchImpl) Ack(ctx context.Context, someID int64) error {
	log.Println("*searchImpl.Ack called", someID)
	return nil
}
