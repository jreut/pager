package main

import (
	"context"
	"database/sql"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sort"
	"strings"
	"time"

	"github.com/jreut/pager/v2/pkg/cli"
	"github.com/jreut/pager/v2/pkg/cmd"
	"github.com/jreut/pager/v2/pkg/global"
	"github.com/jreut/pager/v2/pkg/interval"
	"github.com/jreut/pager/v2/pkg/og"
	"github.com/jreut/pager/v2/pkg/save"
)

var (
	dbpath = "db.sqlite3"
)

func init() {
	if global.Deterministic() {
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
	if global.Deterministic() {
		sort.Strings(names)
	}

	if len(os.Args) <= 1 {
		log.Fatalf("no command given: choose one of %s", names)
	}

	db, err := save.Open(dbpath, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
	defer cancel()

	command, ok := cmds[os.Args[1]]
	if !ok {
		log.Fatalf("unhandled command %q: choose one of %s", os.Args[1], names)
	}
	err = save.WithTx(db, func(tx *sql.Tx) error {
		return command(ctx, os.Args[2:], opts{
			q: save.New(tx),
		})
	})
	if err != nil {
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
	"add-schedule": func(ctx context.Context, args []string, opts opts) error {
		name := flag.String("name", "", "")
		if err := flag.CommandLine.Parse(args); err != nil {
			return err
		}
		return opts.q.AddSchedule(ctx, *name)
	},
	"add-interval": func(ctx context.Context, args []string, opts opts) error {
		times := cli.TimeFlags()
		who := flag.String("who", "", "who")
		kind := flag.String("kind", save.IntervalKindShift, fmt.Sprintf("one of %s", []string{save.IntervalKindShift, save.IntervalKindExclusion}))
		schedule := flag.String("schedule", "", "")
		if err := flag.CommandLine.Parse(args); err != nil {
			return err
		}
		if *who == "" {
			return fmt.Errorf("provide -who")
		}
		if !(*kind == save.IntervalKindExclusion || *kind == save.IntervalKindShift) {
			return fmt.Errorf("provide -kind=%s or -kind=%s", save.IntervalKindShift, save.IntervalKindExclusion)
		}
		start, end, err := times.Times()
		if err != nil {
			return err
		}
		return cmd.AddInterval(ctx, opts.q, save.AddIntervalParams{
			Person:    *who,
			Schedule:  *schedule,
			StartAt:   start,
			EndBefore: end,
			Kind:      *kind,
		})
	},
	"show-schedule": func(ctx context.Context, args []string, opts opts) error {
		schedule := flag.String("schedule", "", "")
		times := cli.TimeFlags()
		if err := flag.CommandLine.Parse(args); err != nil {
			return err
		}
		start, end, err := times.Times()
		if err != nil {
			return err
		}
		out, err := cmd.ShowSchedule(ctx, opts.q, *schedule, start, end)
		if err != nil {
			return err
		}
		return interval.WriteCSV(os.Stdout, out)
	},
	"edit": func(ctx context.Context, args []string, opts opts) error {
		schedule := flag.String("schedule", "", "")
		var actions []cmd.Action
		flag.Var(actionsflag{save.EventKindAdd, &actions}, "add", "")
		flag.Var(actionsflag{save.EventKindRemove, &actions}, "remove", "")
		if err := flag.CommandLine.Parse(args); err != nil {
			return err
		}
		return cmd.EditSchedule(ctx, opts.q, *schedule, actions)
	},
	"generate": func(ctx context.Context, args []string, opts opts) error {
		schedule := flag.String("schedule", "", "")
		times := cli.TimeFlags()
		style := flag.String("style", "", "")
		if err := flag.CommandLine.Parse(args); err != nil {
			return err
		}
		start, end, err := times.Times()
		if err != nil {
			return err
		}
		return cmd.Generate(ctx, opts.q, *schedule, *style, start, end)
	},
	"apply": func(ctx context.Context, args []string, opts opts) error {
		f := flag.String("file", "-", "csv file containing intervals, or stdin if '-'")
		d := flag.String("dst", "stderr", "write to this external destination")
		schedule := flag.String("schedule", "", "")
		if err := flag.CommandLine.Parse(args); err != nil {
			return err
		}
		if *schedule == "" {
			return fmt.Errorf("provide nonempty -schedule")
		}
		r := os.Stdin
		if *f != "-" {
			var err error
			r, err = os.Open(*f)
			if err != nil {
				return err
			}
			defer r.Close()
		}
		var dst cmd.Destination
		switch *d {
		case "opsgenie":
			dst = og.NewHTTPClient(og.DomainDefault, "TODO-KEY")
		case "stderr":
			dst = cmd.FakeDestination{Writer: os.Stderr}
		default:
			return fmt.Errorf("unhandled destination %q", *d)
		}
		return cmd.Apply(ctx, r, dst, *schedule)
	},
}

type actionsflag struct {
	kind string
	val  *[]cmd.Action
}

func (f actionsflag) Set(v string) error {
	before, after, ok := strings.Cut(v, "=")
	if !ok {
		return fmt.Errorf("cannot parse %v: does not contain %q", v, "=")
	}
	at, err := time.Parse(time.RFC3339, after)
	if err != nil {
		return err
	}
	*f.val = append(*f.val, cmd.Action{
		Kind: f.kind,
		Who:  before,
		At:   at,
	})
	return nil
}

func (f actionsflag) String() string {
	return fmt.Sprintf("%s: %s", f.kind, f.val)
}
