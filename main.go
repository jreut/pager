package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/jreut/pager/opsgenie"
	"github.com/jreut/pager/schedule"
)

func main() {
	log.SetFlags(log.Lshortfile | log.LstdFlags)

	if len(os.Args) < 2 {
		log.Fatalf("need a command")
	}
	switch os.Args[1] {
	case "override":
		key := flag.String("key", "", "OpsGenie integration API key")
		flag.CommandLine.Parse(os.Args[2:])
		flag.Parse()

		ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
		defer cancel()
		if err := og(ctx, *key); err != nil {
			log.Fatalf("%+v", err)
		}
		log.Println("ok")
	case "generate":
		generate()
	}

}

func og(ctx context.Context, key string) error {
	client, err := opsgenie.NewClient(key)
	if err != nil {
		return err
	}
	s, err := client.EnsureSchedule(ctx, "reuter test", "Test SRE")
	if err != nil {
		return err
	}
	if err := client.Override(ctx, s, "reuter@cockroachlabs.com", time.Now(), time.Now().AddDate(0, 0, 3)); err != nil {
		return err
	}
	return nil
}

func generate() {
	res := schedule.Builder{
		Interval: schedule.Interval{
			Time:     time.Date(2022, 02, 28, 12, 0, 0, 0, schedule.NYC),
			Duration: 20 * 7 * 24 * time.Hour,
		},
		Next: schedule.MondayFridayShifts,
		Balance: schedule.Balance{
			"carp":    0,
			"darius":  6 * time.Hour,
			"jason":   -4 * 24 * time.Hour,
			"joel":    12 * time.Hour,
			"josh":    -18 * time.Hour,
			"logston": 3 * 24 * time.Hour,
			"reuter":  -3 * 24 * time.Hour,
		},
		Exclusions: []schedule.Exclusion{
			schedule.Exclude("reuter", time.Date(2022, 3, 18, 0, 0, 0, 0, schedule.NYC), time.Date(2022, 3, 22, 0, 0, 0, 0, schedule.NYC)),
			schedule.Exclude("logston", time.Date(2022, 6, 1, 0, 0, 0, 0, time.UTC), time.Date(2022, 7, 1, 0, 0, 0, 0, time.UTC)),
		},
	}.Build()
	log.Printf("%+v", res)
}
