// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Client-side logic for the referee interface.

var websocket;
var foulTeamButton;
var foulRuleButton;
var firstMatchLoad = true;

// // Handles a click on a team button.
// var setFoulTeam = function(teamButton) {
//   if (foulTeamButton) {
//     foulTeamButton.attr("data-selected", false);
//   }
//   foulTeamButton = $(teamButton);
//   foulTeamButton.attr("data-selected", true);

//   $("#commit").prop("disabled", !(foulTeamButton && foulRuleButton));
// };

// // Handles a click on a rule button.
// var setFoulRule = function(ruleButton) {
//   if (foulRuleButton) {
//     foulRuleButton.attr("data-selected", false);
//   }
//   foulRuleButton = $(ruleButton);
//   foulRuleButton.attr("data-selected", true);

//   $("#commit").prop("disabled", !(foulTeamButton && foulRuleButton));
// };



// Sends the foul to the server to add it to the list.
var commitFoul = function(penaltyButton) {
  websocket.send("addFoul", {Alliance: $(penaltyButton).attr("data-alliance"),
      RuleId: parseInt($(penaltyButton).attr("data-rule-id"))});
};

// Removes the foul with the given parameters from the list.
var deleteFoul = function(alliance, ruleId, timeSec) {
  websocket.send("deleteFoul", {Alliance: alliance, RuleId: parseInt(ruleId),
      TimeInMatchSec: timeSec});
};

// Cycles through no card, yellow card, and red card.
var cycleCard = function(cardButton) {
  var newCard = "";
  if ($(cardButton).attr("data-card") === "") {
    newCard = "yellow";
  } else if ($(cardButton).attr("data-card") === "yellow") {
    newCard = "red";
  }
  websocket.send("card", {Alliance: $(cardButton).attr("data-alliance"),
      TeamId: parseInt($(cardButton).attr("data-card-team")), Card: newCard});
  $(cardButton).attr("data-card", newCard);
};

// Sends a websocket message to signal to the volunteers that they may enter the field.
var signalVolunteers = function() {
  websocket.send("signalVolunteers");
};

var signalReview = function() {
  websocket.send("matchReview");
}

// Sends a websocket message to signal to the teams that they may enter the field.
var signalReset = function() {
  websocket.send("signalReset");
};

// Signals the scorekeeper that foul entry is complete for this match.
var commitMatch = function() {
  websocket.send("commitMatch");
  document.querySelectorAll('#foulButtons').forEach(function(el) {
    el.style.display = 'none';
 });
 $("#redScoreButtons").hide()
 $("#blueScoreButtons").hide()
 $("#commitButtons").hide()
 $("#foulList").hide()
};

// Handles a websocket message to update the teams for the current match.
var handleMatchLoad = function(data) {
  // Since the server always sends a matchLoad message upon establishing the websocket connection, ignore the first one.
  if (!firstMatchLoad) {
    location.reload();
  }
  firstMatchLoad = false;
};

$(function() {
  // Activate tooltips above the rule buttons.
  $("[data-toggle=tooltip]").tooltip({"placement": "top"});

  // Set up the websocket back to the server.
  websocket = new CheesyWebsocket("/panels/referee/websocket", {
    matchLoad: function(event) { handleMatchLoad(event.data) },
    matchTime: function(event) { handleMatchTime(event.data); },
    matchTiming: function(event) { handleMatchTiming(event.data); },
    realtimeScore: function(event) { handleRealtimeScore(event.data); },
    scoringStatus: function(event) { handleScoringStatus(event.data); }
  });

  clearFoul();
});


var handleMatchTime = function(data) {
  translateMatchTime(data, function(matchState, matchStateText, countdownSec) {
    var countdownString = String(countdownSec % 60);
    if (countdownString.length === 1) {
      countdownString = "0" + countdownString;
    }
    countdownString = Math.floor(countdownSec / 60) + ":" + countdownString;
    $(".match").attr("data-state", matchState);
    $("#matchState").text(matchStateText);
    $("#matchTime").text(getCountdown(data.MatchState, data.MatchTimeSec));
  });

};

// Handles a websocket message to update the match score.
var handleRealtimeScore = function(data) {
  $("#redScore").text(data.Red.ScoreSummary.Score);
  $("#blueScore").text(data.Blue.ScoreSummary.Score);
};

// Handles a websocket message to signal whether the referee and scorers have committed after the match.
var handleScoringStatus = function(data) {
  scoreIsReady = data.RefereeScoreReady && data.RedScoreReady && data.BlueScoreReady;
  $("#refereeScoreStatus").attr("data-ready", data.RefereeScoreReady);
  $("#redScoreStatus").text("Referee Stands " + data.NumRedScoringPanelsReady + "/" + data.NumRedScoringPanels);
  $("#redScoreStatus").attr("data-ready", data.RedScoreReady);
  $("#blueScoreStatus").text("Blue Scoring " + data.NumBlueScoringPanelsReady + "/" + data.NumBlueScoringPanels);
  $("#blueScoreStatus").attr("data-ready", data.BlueScoreReady);
};

// Handles an element click and sends the appropriate websocket message.
var handleClick = function(shortcut) {
  websocket.send(shortcut);
};