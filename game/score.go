// Copyright 2022 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Model representing the instantaneous score of a match.

package game

type Score struct {
	AutoPoints    int
	TeleopPoints  int
	EndgamePoints int
	Fouls         []Foul
	ElimDq        bool
}

var FoulPointsAwarded = 4
var TechFoulPointsAwarded = 8

var RefereeAutoPointsAwarded = 1
var RefereeTelePointsAwarded = 1
var RefereeEndPointsAwarded = 1

// Represents the state of a robot at the end of the match.
type EndgameStatus int

// Calculates and returns the summary fields used for ranking and display.
func (score *Score) Summarize(opponentFouls []Foul) *ScoreSummary {
	summary := new(ScoreSummary)

	// Leave the score at zero if the team was disqualified.
	if score.ElimDq {
		return summary
	}

	// Calculate penalty points.
	for _, foul := range opponentFouls {
		summary.FoulPoints += foul.PointValue()
	}

	// Check for the opponent fouls that automatically trigger a ranking point.
	// Note: There are no such fouls in the 2022 game; leaving this comment for future years.

	summary.AutoPoints = score.AutoPoints
	summary.TeleopPoints = score.TeleopPoints
	summary.EndgamePoints = score.EndgamePoints
	summary.Score = summary.AutoPoints + summary.TeleopPoints + summary.EndgamePoints + summary.FoulPoints

	return summary

}

// Returns true if and only if all fields of the two scores are equal.
func (score *Score) Equals(other *Score) bool {
	if score.AutoPoints != other.AutoPoints ||
		score.TeleopPoints != other.TeleopPoints ||
		score.EndgamePoints != other.EndgamePoints {
		return false
	}

	for i, foul := range score.Fouls {
		if foul != other.Fouls[i] {
			return false
		}
	}

	return true
}
