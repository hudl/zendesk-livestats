package main

import (
	"fmt"
	"strconv"
	"time"

	"encoding/json"

	"github.com/hudl/zendesk-livestats/config"
	"github.com/hudl/zendesk-livestats/logging"

	"github.com/hudl/ZeGo/zego"
	golog "github.com/op/go-logging"
)

type TicketDetails struct {
	replyTime   int
	createdTime time.Time
}

const dateFmt = "2006-01-02"
const dateZendesk = "Jan-2"

var log = golog.MustGetLogger("main")

func getAverageReplyTime(tickets map[int]TicketDetails, startDate time.Time, endDate time.Time) float64 {
	var ticketsWithValue int = 0
	var replyTimeSum int = 0

	for _, ticket := range tickets {
		if (ticket.createdTime.After(startDate)) && (ticket.createdTime.Before(endDate)) {

			if ticket.replyTime > 0 {
				replyTimeSum += ticket.replyTime
				ticketsWithValue += 1
			}
		}
	}

	return float64(replyTimeSum) / float64(ticketsWithValue)
}

func main() {
	logging.Configure()
	log.Info("Beginning zendesk-livestats...")

	cfg := config.GetConfig()
	zdAuth := zego.Auth{
		cfg.Username,
		cfg.Password,
		cfg.BaseUrl,
	}
	log.Info("cfg.Username " + cfg.Username)

	var allTickets map[int]TicketDetails = make(map[int]TicketDetails)

	for {
		now := time.Now()
		tomorrow := now.Add(24 * time.Hour)
		yesterday := now.Add(-24 * time.Hour)
		monthToDate := now.Add(-30 * 24 * time.Hour)
		query := fmt.Sprintf("created>%s created<%s", monthToDate.Format(dateFmt), tomorrow.Format(dateFmt))
		log.Info("query " + query)

		//Search for tickets between two dates
		searchResponse, err := zdAuth.Search(query)

		if err != nil {
			log.Error("Error while running search: %+v", err)
			// TODO: Decide what we want to do here
		}

		//Get ticket metrics for all tickets
		searchResults := &zego.Search_Results{}
		json.Unmarshal([]byte(searchResponse.Raw), searchResults)

		for _, ticket := range searchResults.Results {
			log.Info("ticket " + string(ticket.Id))
			_, ok := allTickets[ticket.Id]

			//Check if current ticket has been already queried for. If yes, ensure it's got a valid value for replyTime
			if (!ok) || (allTickets[ticket.Id].replyTime < 0) {

				metricsResponse, err := zdAuth.GetTicketMetrics(ticket.Id)
				ticketMetric := &zego.TicketMetric{}
				json.Unmarshal([]byte(metricsResponse.Raw), ticketMetric)

				if err != nil {
					log.Error("Error while running search: %+v", err)
					// TODO: Decide what we want to do here
				}

				str, _ := time.Parse(dateZendesk, ticketMetric.CreatedAt)

				//Add new ticket in map
				var currentValue TicketDetails
				currentValue.createdTime = str
				currentValue.replyTime = ticketMetric.ReplyTime

				allTickets[ticket.Id] = currentValue

				// 10 API hits per minute.
				time.Sleep(7 * time.Second)
			}
		}

		var lastMonthAverage float64 = getAverageReplyTime(allTickets, monthToDate, now)
		var yesterdaysAverage float64 = getAverageReplyTime(allTickets, yesterday, now)
		var todaysAverage float64 = getAverageReplyTime(allTickets, now, tomorrow)

		//Clean-up old tickets
		for id, ticket := range allTickets {

			if ticket.createdTime.Before(monthToDate) {
				delete(allTickets, id)
			}
		}

		//TODO log results.
		log.Info("Month " + strconv.FormatFloat(lastMonthAverage, 'f', -1, 64))
		log.Info("Yesterday " + strconv.FormatFloat(yesterdaysAverage, 'f', -1, 64))
		log.Info("Today " + strconv.FormatFloat(todaysAverage, 'f', -1, 64))

		//time.Sleep(5 * time.Minute)
	}
}
