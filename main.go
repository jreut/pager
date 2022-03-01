package main

import (
	"context"
	"flag"
	"io"
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

		ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
		defer cancel()
		if err := og(ctx, *key); err != nil {
			log.Fatalf("%+v", err)
		}
		log.Println("ok")
	case "generate":
		from := flag.String("from", time.Now().Format(time.RFC3339), "RFC3339-formatted time to start generating the schedule")
		dur := flag.Duration("for", 28*2*24*time.Hour, "length of time to generate")
		flag.CommandLine.Parse(os.Args[2:])
		start, err := time.Parse(time.RFC3339, *from)
		if err != nil {
			log.Fatal(err)
		}

		if err := generate(os.Stdout, schedule.Interval{
			Time: start, Duration: *dur,
		}, schedule.Balance{ // generate the balance from a recorded starting point and then by processing the schedule up to this point
			"carp":    0,
			"darius":  6 * time.Hour,
			"jason":   -4 * 24 * time.Hour,
			"joel":    12 * time.Hour,
			"josh":    -18 * time.Hour,
			"logston": 3 * 24 * time.Hour,
			"reuter":  -3 * 24 * time.Hour,
		}, []schedule.Exclusion{ // read the exclusions from the filesystem
			schedule.Exclude("reuter", time.Date(2022, 3, 18, 0, 0, 0, 0, schedule.NYC), time.Date(2022, 3, 22, 0, 0, 0, 0, schedule.NYC)),
			schedule.Exclude("logston", time.Date(2022, 6, 1, 0, 0, 0, 0, time.UTC), time.Date(2022, 7, 1, 0, 0, 0, 0, time.UTC)),
		}); err != nil {
			log.Fatalf("%+v", err)
		}
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

func generate(w io.Writer, i schedule.Interval, b schedule.Balance, es []schedule.Exclusion) error {
	res := schedule.Builder{
		Interval:   i,
		Next:       schedule.MondayFridayShifts,
		Balance:    b,
		Exclusions: es,
	}.Build()
	return res.Schedule.WriteCSV(w)
}
