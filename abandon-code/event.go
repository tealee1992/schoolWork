
package main

import (
"encoding/json"
"fmt"
"io"
"log"

"github.com/docker/engine-api/client"
"github.com/docker/engine-api/types"
"github.com/docker/engine-api/types/events"
"golang.org/x/net/context"
)

func main() {

	defaultHeaders := map[string]string{"User-Agent": "engine-api-cli-1.0"}
	cli, err := client.NewClient("http://11.0.57.2:2375", "v1.23", nil, defaultHeaders)
	if err != nil {
		panic(err)
	}

	options := types.ContainerListOptions{All: true}
	containers, err := cli.ContainerList(context.Background(), options)
	if err != nil {
		panic(err)
	}

	for _, c := range containers {
		fmt.Println(c.ID)
	}

	body, err := cli.Events(context.Background(), types.EventsOptions{})
	if err != nil {
		log.Fatal(err)
	}

	dec := json.NewDecoder(body)
	for {
		var event events.Message
		err := dec.Decode(&event)
		if err != nil && err == io.EOF {
			break
		}

		log.Println(event)
	}
}
