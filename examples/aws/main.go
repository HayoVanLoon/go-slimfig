package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/HayoVanLoon/go-slimfig"
	aws "github.com/HayoVanLoon/go-slimfig/resolver/aws/secretsmngr"
	"github.com/HayoVanLoon/go-slimfig/resolver/json"
)

func initConfig(references []string) {
	ctx := context.Background()

	secretsManager, err := aws.JSONResolver(ctx)
	if err != nil {
		log.Fatal(err)
	}

	slimfig.SetResolvers(secretsManager, json.Resolver())
	if err = slimfig.Load(ctx, "EX", references...); err != nil {
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
