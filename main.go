package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/jreut/pager/opsgenie"
	"github.com/opsgenie/opsgenie-go-sdk-v2/client"
)

var cfg client.Config

func main() {
	log.SetFlags(log.Lshortfile | log.LstdFlags)
	key := flag.String("key", "", "OpsGenie integration API key")
	flag.Parse()
	cfg.ApiKey = *key

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()
	if err := run(ctx, *key); err != nil {
		log.Fatalf("%+v", err)
	}
	log.Println("ok")
}

func run(ctx context.Context, key string) error {
	client, err := opsgenie.NewClient(key)
	if err != nil {
		return err
	}
	s, err := client.EnsureSchedule(ctx)
	if err != nil {
		return err
	}
	if err := client.Override(ctx, s, "reuter@cockroachlabs.com", time.Now(), time.Now().AddDate(0, 0, 3)); err != nil {
		return err
	}
	return nil
}
