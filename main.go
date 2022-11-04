package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sort"
	"time"

	"github.com/jreut/pager/v2/pkg/cmd"
	"github.com/jreut/pager/v2/pkg/save"
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
	var names []string
	for k := range cmds {
		names = append(names, k)
	}
	sort.Strings(names) // sort for test determinism

	if len(os.Args) <= 1 {
		log.Fatalf("no command given: choose one of %s", names)
	}

	db, err := save.Open(dbpath, nil)
	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
	defer cancel()

	command, ok := cmds[os.Args[1]]
	if !ok {
		log.Fatalf("unhandled command %q: choose one of %s", os.Args[1], names)
	}
	if err := command(ctx, os.Args[2:], opts{
		save.New(db),
	}); err != nil {
		log.Println(err)
		switch {
		case errors.Is(err, cmd.ErrConflict):
			os.Exit(17)
		default:
			os.Exit(1)
		}
	}
	log.Println("ok")
}

type opts struct{ q *save.Queries }

var cmds = map[string]func(context.Context, []string, opts) error{
	"add-person": func(ctx context.Context, args []string, opts opts) error {
		who := flag.String("who", "", "")
		if err := flag.CommandLine.Parse(args); err != nil {
			return err
		}
		if *who == "" {
			return fmt.Errorf("provide -who")
		}
		return opts.q.AddPerson(ctx, *who)
	},
	"add-interval": func(ctx context.Context, args []string, opts opts) error {
		var (
			zero  time.Time
			start time.Time
			end   time.Time
		)
		flag.Var(timeflag{&start}, "start", "start (inclusive)")
		flag.Var(timeflag{&end}, "end", "end (exclusive)")
		dur := flag.Duration("for", 0, "duration")
		who := flag.String("who", "", "who")
		kind := flag.String("kind", save.IntervalKindShift, fmt.Sprintf("one of %s", []string{save.IntervalKindShift, save.IntervalKindExclusion}))
		if err := flag.CommandLine.Parse(args); err != nil {
			return err
		}
		if start == zero {
			return fmt.Errorf("provide -start")
		}
		if *who == "" {
			return fmt.Errorf("provide -who")
		}
		if !(*kind == save.IntervalKindExclusion || *kind == save.IntervalKindShift) {
			return fmt.Errorf("provide -kind=%s or -kind=%s", save.IntervalKindShift, save.IntervalKindExclusion)
		}
		if end == zero && *dur == 0 {
			return fmt.Errorf("provide one of -end or -for")
		}
		if *dur != 0 {
			end = start.Add(*dur)
		}
		return cmd.AddInterval(ctx, opts.q, save.AddIntervalParams{
			Person:    *who,
			StartAt:   start,
			EndBefore: end,
			Kind:      *kind,
		})
	},
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
