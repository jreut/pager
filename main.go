package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"

	"github.com/opsgenie/opsgenie-go-sdk-v2/client"
	"github.com/opsgenie/opsgenie-go-sdk-v2/schedule"
)

func main() {
	key := flag.String("key", "", "OpsGenie integration API key")
	flag.Parse()
	client, err := schedule.NewClient(&client.Config{ApiKey: *key})
	if err != nil {
		panic(err)
	}
	ctx := context.TODO()
	expand := true
	out, err := client.List(ctx, &schedule.ListRequest{Expand: &expand})
	if err != nil {
		panic(err)
	}
	buf, err := json.Marshal(out)
	fmt.Printf("%s", buf)
}
