package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/HayoVanLoon/go-slimfig"
	"github.com/HayoVanLoon/go-slimfig/resolver/json"
	"github.com/HayoVanLoon/go-slimfig/resolver/yaml"
)

func initConfig(references []string) {
	ctx := context.Background()

	slimfig.SetResolvers(yaml.Resolver(), json.Resolver())
	if err := slimfig.Load(ctx, "EX", references...); err != nil {
		log.Fatal(err)
	}
}

func main() {
	initConfig(os.Args[1:])

	s, err := slimfig.JSON()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(s)
}
