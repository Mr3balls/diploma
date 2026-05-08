package handler

import (
	"esports-backend/internal/entity"
	"esports-backend/internal/service"
)

func serviceRegisterInput(req registerRequest) service.RegisterInput {
	return service.RegisterInput{FirstName: req.FirstName, LastName: req.LastName, Email: req.Email, Phone: req.Phone, Nickname: req.Nickname, Password: req.Password}
}

func serviceLoginInput(req loginRequest, ua, ip *string) service.LoginInput {
	return service.LoginInput{Email: req.Email, Password: req.Password, UserAgent: ua, IPAddress: ip}
}

func toCreateTournamentInput(t *entity.Tournament) service.CreateTournamentInput {
	return service.CreateTournamentInput{
		Title:                t.Title,
		Discipline:           t.Discipline,
		Description:          t.Description,
		Rules:                t.Rules,
		Location:             t.Location,
		MaxTeams:             t.MaxTeams,
		Format:               t.Format,
		GroupCount:           t.GroupCount,
		RegistrationDeadline: t.RegistrationDeadline,
		StartAt:              t.StartAt,
		Visibility:           t.Visibility,
	}
}
