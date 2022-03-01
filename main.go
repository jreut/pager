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
		balpath := flag.String("balance", "data/balance.csv", "path to balance file")
		espath := flag.String("exclusions", "data/exclusions.csv", "path to exclusions file")
		flag.CommandLine.Parse(os.Args[2:])
		start, err := time.Parse(time.RFC3339, *from)
		if err != nil {
			log.Fatal(err)
		}

		bal, err := os.Open(*balpath)
		if err != nil {
			log.Fatal(err)
		}
		defer bal.Close()
		es, err := os.Open(*espath)
		if err != nil {
			log.Fatal(err)
		}
		defer bal.Close()

		if err := generate(os.Stdout, schedule.Interval{
			Time: start, Duration: *dur,
		}, bal, es); err != nil {
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

func generate(w io.Writer, i schedule.Interval, balance, exclusions io.Reader) error {
	es, err := schedule.ExclusionsFromCSV(exclusions)
	if err != nil {
		return err
	}
	b, err := schedule.BalanceFromCSV(balance)
	if err != nil {
		return err
	}
	res := schedule.Builder{
		Interval:   i,
		Next:       schedule.MondayFridayShifts,
		Balance:    b,
		Exclusions: es,
	}.Build()
	return res.Schedule.WriteCSV(w)
}
