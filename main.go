package main

import (
	"context"
	"database/sql"
	"encoding/csv"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sort"
	"strings"
	"time"

	"github.com/jreut/pager/v2/pkg/cmd"
	"github.com/jreut/pager/v2/pkg/global"
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
	sort.Strings(names) // sort for test determinism

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
		times := addtimeflags()
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
		times := addtimeflags()
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
		w := csv.NewWriter(os.Stdout)
		defer w.Flush()
		if err := w.Write([]string{"start_at", "end_before", "person"}); err != nil {
			return err
		}
		for _, i := range out {
			if err := w.Write([]string{
				i.StartAt.Format(time.RFC3339),
				i.EndBefore.Format(time.RFC3339),
				i.Person,
			}); err != nil {
				return err
			}
		}
		return nil
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
		times := addtimeflags()
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

func addtimeflags() *timeflags {
	var out timeflags
	flag.Var(timeflag{&out.start}, "start", "start (inclusive)")
	flag.Var(timeflag{&out.end}, "end", "end (exclusive)")
	flag.DurationVar(&out.dur, "for", 0, "duration")
	return &out
}

type timeflags struct {
	start, end time.Time
	dur        time.Duration
}

func (f timeflags) Times() (start, end time.Time, err error) {
	var zero time.Time
	if f.start == zero {
		return zero, zero, fmt.Errorf("provide -start")
	}
	if (f.end == zero) == (f.dur == 0) {
		return zero, zero, fmt.Errorf("provide one of -end or -for")
	}
	if f.dur != 0 {
		return f.start, f.start.Add(f.dur), nil
	}
	return f.start, f.end, nil
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
