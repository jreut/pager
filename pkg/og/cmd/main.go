package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/url"
	"os"
	"os/signal"
	"strings"

	"github.com/jreut/pager/v2/pkg/cli"
	"github.com/jreut/pager/v2/pkg/interval"
	"github.com/jreut/pager/v2/pkg/og"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
	defer cancel()

	if err := run(ctx); err != nil {
		log.Fatal(err)
	}
	log.Println("ok")
}

func run(ctx context.Context) error {
	key := os.Getenv("OPSGENIE_KEY")
	debug := flag.Bool("debug", false, "")

	switch os.Args[1] {
	case "get-timeline":
		schedule := flag.String("s", "", "schedule id")
		times := cli.TimeFlags()
		if err := flag.CommandLine.Parse(os.Args[2:]); err != nil {
			return err
		}
		from, to, err := times.Times()
		if err != nil {
			return err
		}
		client := og.NewHTTPClient(og.DomainDefault, key, *debug)
		out, err := client.GetTimeline(ctx, *schedule, from, to)
		if err != nil {
			return err
		}
		return interval.WriteCSV(os.Stdout, out)
	case "http":
		method := flag.String("m", "GET", "HTTP method")
		path := flag.String("p", "", "request path")
		body := flag.String("b", "", "JSON-encoded request body")
		if err := flag.CommandLine.Parse(os.Args[2:]); err != nil {
			return err
		}
		client := og.NewHTTPClient(og.DomainDefault, key, *debug)
		switch strings.ToUpper(*method) {
		case "GET":
			var query url.Values
			if *body != "" {
				if err := json.Unmarshal([]byte(*body), &query); err != nil {
					return fmt.Errorf("unmarshaling query parameters: %w", err)
				}
			}
			res, err := client.Get(ctx, *path, query)
			if err != nil {
				return fmt.Errorf("HTTP GET: %w", err)
			}
			log.Println(res.Status)
			defer res.Body.Close()
			_, err = io.Copy(os.Stdout, res.Body)
			return err
		case "POST":
			var data interface{}
			if *body != "" {
				if err := json.Unmarshal([]byte(*body), &data); err != nil {
					return fmt.Errorf("unmarshaling request body: %w", err)
				}
			}
			res, err := client.Post(ctx, *path, data)
			if err != nil {
				return fmt.Errorf("HTTP POST: %w", err)
			}
			log.Println(res.Status)
			defer res.Body.Close()
			_, err = io.Copy(os.Stdout, res.Body)
			return err
		default:
			return fmt.Errorf("unhandled method %s", *method)
		}
	default:
		return fmt.Errorf("unhandled command %s", os.Args[1])
	}
}
