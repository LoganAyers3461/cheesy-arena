// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Web handlers for scoring interface.

package web

import (
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/Team254/cheesy-arena/field"
	"github.com/Team254/cheesy-arena/game"
	"github.com/Team254/cheesy-arena/model"
	"github.com/Team254/cheesy-arena/websocket"
	"github.com/gorilla/mux"
	"github.com/mitchellh/mapstructure"
)

// Renders the scoring interface which enables input of scores in real-time.
func (web *Web) scoringPanelHandler(w http.ResponseWriter, r *http.Request) {
	if !web.userIsAdmin(w, r) {
		return
	}

	vars := mux.Vars(r)
	alliance := vars["alliance"]
	if alliance != "red" {
		// handleWebErr(w, fmt.Errorf("Invalid alliance '%s'.", alliance))
		alliance = "red"

	}

	template, err := web.parseFiles("templates/scoring_panel.html")
	if err != nil {
		handleWebErr(w, err)
		return
	}

	match := web.arena.CurrentMatch
	matchType := match.CapitalizedType()
	red1 := web.arena.AllianceStations["R1"].Team
	if red1 == nil {
		red1 = &model.Team{}
	}
	red2 := web.arena.AllianceStations["R2"].Team
	if red2 == nil {
		red2 = &model.Team{}
	}
	red3 := web.arena.AllianceStations["R3"].Team
	if red3 == nil {
		red3 = &model.Team{}
	}
	blue1 := web.arena.AllianceStations["B1"].Team
	if blue1 == nil {
		blue1 = &model.Team{}
	}
	blue2 := web.arena.AllianceStations["B2"].Team
	if blue2 == nil {
		blue2 = &model.Team{}
	}
	blue3 := web.arena.AllianceStations["B3"].Team
	if blue3 == nil {
		blue3 = &model.Team{}
	}
	data := struct {
		*model.EventSettings
		MatchType        string
		MatchDisplayName string
		Red1             *model.Team
		Red2             *model.Team
		Red3             *model.Team
		Blue1            *model.Team
		Blue2            *model.Team
		Blue3            *model.Team
		RedFouls         []game.Foul
		BlueFouls        []game.Foul
		RedCards         map[string]string
		BlueCards        map[string]string
		Rules            map[int]*game.Rule
		Alliance         string
	}{web.arena.EventSettings, matchType, match.DisplayName, red1, red2, red3, blue1, blue2, blue3,
		web.arena.RedRealtimeScore.CurrentScore.Fouls, web.arena.BlueRealtimeScore.CurrentScore.Fouls,
		web.arena.RedRealtimeScore.Cards, web.arena.BlueRealtimeScore.Cards, game.GetAllRules(), alliance}
	err = template.ExecuteTemplate(w, "scoring_panel.html", data)
	if err != nil {
		handleWebErr(w, err)
		return
	}
}

// The websocket endpoint for the scoring interface client to send control commands and receive status updates.
func (web *Web) scoringPanelWebsocketHandler(w http.ResponseWriter, r *http.Request) {
	if !web.userIsAdmin(w, r) {
		return
	}

	vars := mux.Vars(r)
	alliance := vars["alliance"]
	if alliance != "red" {
		// handleWebErr(w, fmt.Errorf("Invalid alliance '%s'.", alliance))
		// return
		alliance = "red"
	}

	ws, err := websocket.NewWebsocket(w, r)
	if err != nil {
		handleWebErr(w, err)
		return
	}
	defer ws.Close()
	web.arena.ScoringPanelRegistry.RegisterPanel(alliance, ws)
	web.arena.ScoringStatusNotifier.Notify()
	defer web.arena.ScoringStatusNotifier.Notify()
	defer web.arena.ScoringPanelRegistry.UnregisterPanel(alliance, ws)

	// Subscribe the websocket to the notifiers whose messages will be passed on to the client, in a separate goroutine.
	go ws.HandleNotifiers(web.arena.MatchLoadNotifier, web.arena.ReloadDisplaysNotifier, web.arena.MatchTimingNotifier, web.arena.ArenaStatusNotifier, web.arena.MatchTimeNotifier, web.arena.RealtimeScoreNotifier)

	// Loop, waiting for commands and responding to them, until the client closes the connection.
	for {
		command, data, err := ws.Read()
		if err != nil {
			if err == io.EOF {
				// Client has closed the connection; nothing to do here.
				return
			}
			log.Println(err)
			return
		}

		// score := &(*realtimeScore).CurrentScore
		// scoreChanged := false

		if command == "addFoul" {
			args := struct {
				Alliance string
				TeamId   int
				RuleId   int
			}{}
			err = mapstructure.Decode(data, &args)
			if err != nil {
				ws.WriteError(err.Error())
				continue
			}

			// Add the foul to the correct alliance's list.
			foul := game.Foul{RuleId: args.RuleId, TimeInMatchSec: web.arena.MatchTimeSec()}
			if args.Alliance == "red" {
				web.arena.RedRealtimeScore.CurrentScore.Fouls =
					append(web.arena.RedRealtimeScore.CurrentScore.Fouls, foul)
			} else {
				web.arena.BlueRealtimeScore.CurrentScore.Fouls =
					append(web.arena.BlueRealtimeScore.CurrentScore.Fouls, foul)
			}
			web.arena.RealtimeScoreNotifier.Notify()
		}

		if command == "commitMatch" {
			if web.arena.MatchState != field.PostMatch {
				// Don't allow committing the score until the match is over.
				ws.WriteError("Cannot commit score: Match is not over.")
				continue
			}
			web.arena.ScoringPanelRegistry.SetScoreCommitted(alliance, ws)
			web.arena.ScoringStatusNotifier.Notify()
		} else if !web.arena.Plc.IsEnabled() {
			switch strings.ToUpper(command) {
			case "Q":
				web.arena.RedRealtimeScore.CurrentScore.AutoPoints -= game.RefereeAutoPointsAwarded

				web.arena.RealtimeScoreNotifier.Notify()
			case "A":
				web.arena.RedRealtimeScore.CurrentScore.TeleopPoints -= game.RefereeTelePointsAwarded
				web.arena.RealtimeScoreNotifier.Notify()
			case "W":
				web.arena.RedRealtimeScore.CurrentScore.AutoPoints += game.RefereeAutoPointsAwarded

				web.arena.RealtimeScoreNotifier.Notify()
			case "S":
				web.arena.RedRealtimeScore.CurrentScore.TeleopPoints += game.RefereeTelePointsAwarded

				web.arena.RealtimeScoreNotifier.Notify()
			case "T":
				web.arena.BlueRealtimeScore.CurrentScore.AutoPoints -= game.RefereeAutoPointsAwarded
				web.arena.RealtimeScoreNotifier.Notify()
			case "G":
				web.arena.BlueRealtimeScore.CurrentScore.TeleopPoints += game.RefereeTelePointsAwarded
				web.arena.RealtimeScoreNotifier.Notify()
			case "Y":
				web.arena.BlueRealtimeScore.CurrentScore.AutoPoints += game.RefereeAutoPointsAwarded
				web.arena.RealtimeScoreNotifier.Notify()
			case "H":
				web.arena.BlueRealtimeScore.CurrentScore.TeleopPoints += game.RefereeTelePointsAwarded
				web.arena.RealtimeScoreNotifier.Notify()

			case "N":
				web.arena.BlueRealtimeScore.CurrentScore.EndgamePoints += game.RefereeEndPointsAwarded
				web.arena.RealtimeScoreNotifier.Notify()

			case "B":
				web.arena.BlueRealtimeScore.CurrentScore.EndgamePoints -= game.RefereeEndPointsAwarded
				web.arena.RealtimeScoreNotifier.Notify()

			case "X":
				web.arena.RedRealtimeScore.CurrentScore.EndgamePoints += game.RefereeEndPointsAwarded

				web.arena.RealtimeScoreNotifier.Notify()

			case "Z":
				web.arena.RedRealtimeScore.CurrentScore.EndgamePoints -= game.RefereeEndPointsAwarded

				web.arena.RealtimeScoreNotifier.Notify()

			}

		}

	}
}

// Increments the cargo count for the given goal.
func incrementGoal(goal int) {
	// Use just the first hub quadrant for manual scoring.
	goal++
}

func decrementGoal(goal int) {
	// Use just the first hub quadrant for manual scoring.
	if goal > 0 {
		goal--
	}

}
