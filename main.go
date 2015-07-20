package main

import (
	"fmt"
	"time"

	"github.com/hudl/zendesk-livestats/config"
	"github.com/hudl/zendesk-livestats/logging"

	"github.com/hudl/ZeGo/zego"
	golog "github.com/op/go-logging"
)

const dateFmt = "2006-01-02"

var log = golog.MustGetLogger("main")

func main() {
	logging.Configure()
	log.Info("Beginning zendesk-livestats...")

	cfg := config.GetConfig()
	zdAuth := zego.Auth{
		cfg.Username,
		cfg.Password,
		cfg.BaseUrl,
	}

	now := time.Now()
	tomorrow := now.Add(24 * time.Hour).Format(dateFmt)
	monthToDate := now.Add(-30 * 24 * time.Hour).Format(dateFmt)
	query := fmt.Sprintf("created>%s created<%s", monthToDate, tomorrow)
	res, err := zdAuth.Search(query)
	if err != nil {
		log.Error("Error while running search: %+v", err)
		// TODO: Decide what we want to do here
	}

	// Three search queries - created today, yesterday, and past month
	// For each of them loop through and write to logs

}
