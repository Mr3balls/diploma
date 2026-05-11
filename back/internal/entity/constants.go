package entity

const (
	PlatformRolePlayer        = "player"
	PlatformRolePlatformAdmin = "platform_admin"
)

const (
	TournamentRoleOwner   = "owner"
	TournamentRoleManager = "manager"
)

const (
	TournamentStatusDraft              = "draft"
	TournamentStatusRegistrationOpen   = "registration_open"
	TournamentStatusRegistrationClosed = "registration_closed"
	TournamentStatusBracketGenerated   = "bracket_generated"
	TournamentStatusInProgress         = "in_progress"
	TournamentStatusFinished           = "finished"
	TournamentStatusCancelled          = "cancelled"
	// Challonge-style lifecycle statuses
	TournamentStatusReady     = "ready"
	TournamentStatusCompleted = "completed"
)

const (
	TournamentRoleOrganizer   = "organizer"
	TournamentRoleCoOrganizer = "co_organizer"
	TournamentRoleParticipant = "participant"
	TournamentRoleViewer      = "viewer"
)

const (
	ParticipantStatusActive     = "active"
	ParticipantStatusEliminated = "eliminated"
	ParticipantStatusChampion   = "champion"
)

const (
	ResultReportPending  = "pending"
	ResultReportApproved = "approved"
	ResultReportRejected = "rejected"
)

const (
	InviteStatusPending  = "pending"
	InviteStatusAccepted = "accepted"
	InviteStatusRejected = "rejected"
	InviteStatusExpired  = "expired"
)

const (
	TournamentVisibilityPublic  = "public"
	TournamentVisibilityPrivate = "private"
)

const (
	TeamStatusPending              = "pending"
	TeamStatusAwaitingConfirmation = "awaiting_confirmation"
	TeamStatusReadyForReview       = "ready_for_review"
	TeamStatusApproved             = "approved"
	TeamStatusRejected             = "rejected"
)

const (
	MemberConfirmationFound               = "found"
	MemberConfirmationNotFound            = "not_found"
	MemberConfirmationPendingConfirmation = "pending_confirmation"
	MemberConfirmationConfirmed           = "confirmed"
	MemberConfirmationDeclined            = "declined"
	MemberConfirmationRemoved             = "removed"
)

const (
	ImportBatchStatusPending      = "pending"
	ImportBatchStatusParsing      = "parsing"
	ImportBatchStatusPreviewReady = "preview_ready"
	ImportBatchStatusConfirmed    = "confirmed"
	ImportBatchStatusFailed       = "failed"
)

const (
	ImportRowStatusNew         = "new"
	ImportRowStatusValid       = "valid"
	ImportRowStatusNeedsReview = "needs_review"
	ImportRowStatusDuplicate   = "duplicate"
	ImportRowStatusRejected    = "rejected"
	ImportRowStatusConfirmed   = "confirmed"
)

const (
	BracketSectionWB = "WB"
	BracketSectionLB = "LB"
	BracketSectionGF = "GF"
)

const (
	MatchStatusScheduled            = "scheduled"
	MatchStatusAwaitingConfirmation = "awaiting_confirmation"
	MatchStatusConfirmed            = "confirmed"
	MatchStatusRescheduleRequested  = "reschedule_requested"
	MatchStatusIssueReported        = "issue_reported"
	MatchStatusInProgress           = "in_progress"
	MatchStatusFinished             = "finished"
	MatchStatusCancelled            = "cancelled"
)

const (
	MatchTeamConfirmationPending             = "pending"
	MatchTeamConfirmationReadyConfirmed      = "ready_confirmed"
	MatchTeamConfirmationRescheduleRequested = "reschedule_requested"
	MatchTeamConfirmationIssueReported       = "issue_reported"
)

const (
	NotificationAddedToTeam              = "added_to_team"
	NotificationTeamParticipationConfirm = "team_participation_confirmed"
	NotificationTeamParticipationDecline = "team_participation_declined"
	NotificationMatchAssigned            = "match_assigned"
	NotificationMatchTimeChanged         = "match_time_changed"
	NotificationMatchRescheduled         = "match_rescheduled"
	NotificationMatchCancelled           = "match_cancelled"
	NotificationResultSubmitted          = "result_submitted"
	NotificationResultConfirmed          = "result_confirmed"
	NotificationTournamentFinished       = "tournament_finished"
)

const (
	MemberRoleCaptain    = "captain"
	MemberRolePlayer     = "player"
	MemberRoleSubstitute = "substitute"
)
