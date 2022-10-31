package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/jreut/pager/v2/internal/save"
)

var (
	deterministic bool
	dbpath        = "db.sqlite3"
)

func init() {
	if os.Getenv("DETERMINISTIC") == "1" {
		deterministic = true
	}
	if deterministic {
		log.SetFlags(log.Lshortfile)
	} else {
		log.SetFlags(log.LstdFlags | log.Lshortfile)
	}
	if v, ok := os.LookupEnv("DB"); ok {
		dbpath = v
	}
}

func main() {
	if len(os.Args) <= 1 {
		log.Fatal("usage: TODO")
	}

	db, err := save.Open(dbpath, nil)
	if err != nil {
		log.Fatal(err)
	}
	q := save.New(db)

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
	defer cancel()

	switch mode := os.Args[1]; mode {
	case "add-person":
		who := flag.String("who", "", "")
		if err := flag.CommandLine.Parse(os.Args[2:]); err != nil {
			log.Fatal(err)
		}
		if *who == "" {
			log.Fatal("provide -who")
		}
		if err := addperson(ctx, q, *who); err != nil {
			log.Fatal(err)
		}
	case "add-shift":
		var (
			empty time.Time
			start time.Time
			end   time.Time
		)
		flag.Var(timeflag{&start}, "start", "start (inclusive)")
		flag.Var(timeflag{&end}, "end", "end (exclusive)")
		dur := flag.Duration("for", 0, "duration")
		who := flag.String("who", "", "who")
		if err := flag.CommandLine.Parse(os.Args[2:]); err != nil {
			log.Fatal(err)
		}
		if start == empty {
			log.Fatal("provide -start")
		}
		if *who == "" {
			log.Fatal("provide -who")
		}
		if end == empty && *dur == 0 {
			log.Fatal("provide one of -end or -for")
		}
		if *dur != 0 {
			end = start.Add(*dur)
		}
		if err := addshift(ctx, q, *who, start, end); err != nil {
			log.Fatal(err)
		}

	/*
		case "apply":
			var (
				start time.Time
				end   time.Time
			)
			flag.Var(&timeflag{start}, "start", "start (inclusive)")
			flag.Var(&timeflag{end}, "end", "end (exclusive)")
			flag.CommandLine.Parse(os.Args[2:])
		case "report":
			var (
				start time.Time
				end   time.Time
			)
			flag.Var(&timeflag{start}, "start", "start (inclusive)")
			flag.Var(&timeflag{end}, "end", "end (exclusive)")
			who := flag.String("person", "", "who")
			flag.CommandLine.Parse(os.Args[2:])
		case "who":
			var at time.Time
			flag.Var(&timeflag{at}, "at", "")
			flag.CommandLine.Parse(os.Args[2:])
	*/
	default:
		log.Fatalf("unhandled mode %q: TODO", mode)
	}
}

type timeflag struct{ *time.Time }

func (f timeflag) Set(v string) error {
	out, err := time.Parse(time.RFC3339, v)
	if err != nil {
		return err
	}
	*f.Time = out
	return nil
}

func (f timeflag) String() string {
	if f.Time == nil {
		return "<nil>"
	}
	return f.Format(time.RFC3339)
}

func addshift(ctx context.Context, q *save.Queries, who string, start, end time.Time) error {
	return q.AddShift(ctx, save.AddShiftParams{
		Person:    who,
		StartAt:   start,
		EndBefore: end,
	})
}
func addperson(ctx context.Context, q *save.Queries, who string) error {
	return q.AddPerson(ctx, who)
}
