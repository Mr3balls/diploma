package service

import (
	"testing"

	"esports-backend/internal/entity"
)

func TestIsTeamReadyForReview(t *testing.T) {
	members := []entity.TeamMember{
		{IsCaptain: true, MemberRole: entity.MemberRoleCaptain, ConfirmationStatus: entity.MemberConfirmationConfirmed},
		{MemberRole: entity.MemberRolePlayer, ConfirmationStatus: entity.MemberConfirmationConfirmed},
		{MemberRole: entity.MemberRolePlayer, ConfirmationStatus: entity.MemberConfirmationConfirmed},
		{MemberRole: entity.MemberRolePlayer, ConfirmationStatus: entity.MemberConfirmationConfirmed},
		{MemberRole: entity.MemberRolePlayer, ConfirmationStatus: entity.MemberConfirmationPendingConfirmation},
	}
	if !IsTeamReadyForReview(members) {
		t.Fatalf("expected team to be ready when captain + 4 players are confirmed")
	}
}

func TestIsTeamReadyForReviewFailsWithoutCaptain(t *testing.T) {
	members := []entity.TeamMember{
		{IsCaptain: true, MemberRole: entity.MemberRoleCaptain, ConfirmationStatus: entity.MemberConfirmationPendingConfirmation},
		{MemberRole: entity.MemberRolePlayer, ConfirmationStatus: entity.MemberConfirmationConfirmed},
		{MemberRole: entity.MemberRolePlayer, ConfirmationStatus: entity.MemberConfirmationConfirmed},
		{MemberRole: entity.MemberRolePlayer, ConfirmationStatus: entity.MemberConfirmationConfirmed},
		{MemberRole: entity.MemberRolePlayer, ConfirmationStatus: entity.MemberConfirmationConfirmed},
	}
	if IsTeamReadyForReview(members) {
		t.Fatalf("expected team not to be ready when captain is not confirmed")
	}
}
