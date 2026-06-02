// Package notif provides localized notification texts for ru/en/kk.
package notif

import "fmt"

// Texts holds a localized title and message for a notification.
type Texts struct {
	Title   string
	Message string
}

func AddedToTeam(lang, teamName string) Texts {
	switch lang {
	case "en":
		return Texts{"Added to a team", fmt.Sprintf("Confirm your participation in team «%s»", teamName)}
	case "kk":
		return Texts{"Командаға қосылдыңыз", fmt.Sprintf("«%s» командасындағы қатысуды растаңыз", teamName)}
	default:
		return Texts{"Вас добавили в команду", fmt.Sprintf("Подтвердите участие в команде «%s»", teamName)}
	}
}

func TeamApproved(lang, teamName string) Texts {
	switch lang {
	case "en":
		return Texts{"Team approved", fmt.Sprintf("Your team «%s» has been approved for the tournament.", teamName)}
	case "kk":
		return Texts{"Команда мақұлданды", fmt.Sprintf("«%s» командаңыз турнирге мақұлданды.", teamName)}
	default:
		return Texts{"Команда одобрена", fmt.Sprintf("Ваша команда «%s» подтверждена и допущена к турниру.", teamName)}
	}
}

func TeamDeclined(lang, teamName, reason string) Texts {
	switch lang {
	case "en":
		return Texts{"Team rejected", fmt.Sprintf("Your team «%s» was rejected. Reason: %s", teamName, reason)}
	case "kk":
		return Texts{"Команда қабылданбады", fmt.Sprintf("«%s» командаңыз қабылданбады. Себеп: %s", teamName, reason)}
	default:
		return Texts{"Команда отклонена", fmt.Sprintf("Ваша команда «%s» была отклонена. Причина: %s", teamName, reason)}
	}
}

func MatchAssigned(lang string) Texts {
	switch lang {
	case "en":
		return Texts{"Match assigned", "The bracket has been generated. Your team has received an opponent."}
	case "kk":
		return Texts{"Матч тағайындалды", "Сетка жасалды. Командаңыздың қарсыласы белгіленді."}
	default:
		return Texts{"Матч назначен", "Сетка турнира сформирована. Ваша команда получила соперника."}
	}
}

func MatchTimeChanged(lang string) Texts {
	switch lang {
	case "en":
		return Texts{"Match time updated", "The manager has updated the match details."}
	case "kk":
		return Texts{"Матч уақыты жаңартылды", "Менеджер матч деректерін жаңартты."}
	default:
		return Texts{"Время матча обновлено", "Менеджер обновил данные матча."}
	}
}

func MatchRescheduled(lang string) Texts {
	switch lang {
	case "en":
		return Texts{"Reschedule requested", "A team has requested to reschedule the match."}
	case "kk":
		return Texts{"Ауыстыру сұралды", "Команда матчты ауыстыруды сұрады."}
	default:
		return Texts{"Запрошен перенос матча", "Команда запросила перенос матча."}
	}
}

func MatchCancelled(lang string) Texts {
	switch lang {
	case "en":
		return Texts{"Issue reported", "A problem has been reported for the match."}
	case "kk":
		return Texts{"Мәселе хабарланды", "Матч бойынша мәселе хабарланды."}
	default:
		return Texts{"По матчу сообщено о проблеме", "По матчу сообщено о проблеме."}
	}
}

func ResultSubmitted(lang string) Texts {
	switch lang {
	case "en":
		return Texts{"Result submitted", "A match result has been submitted for confirmation."}
	case "kk":
		return Texts{"Нәтиже жіберілді", "Матч нәтижесі растауға жіберілді."}
	default:
		return Texts{"Результат матча отправлен", "По матчу отправлен результат на подтверждение."}
	}
}

func ResultConfirmed(lang string) Texts {
	switch lang {
	case "en":
		return Texts{"Result confirmed", "The manager has confirmed the match result."}
	case "kk":
		return Texts{"Нәтиже расталды", "Менеджер матч нәтижесін растады."}
	default:
		return Texts{"Результат матча подтверждён", "Менеджер подтвердил результат матча."}
	}
}

func TournamentFinished(lang string) Texts {
	switch lang {
	case "en":
		return Texts{"Tournament finished", "The tournament has ended. The final result has been recorded."}
	case "kk":
		return Texts{"Турнир аяқталды", "Турнир аяқталды. Қорытынды нәтиже тіркелді."}
	default:
		return Texts{"Турнир завершён", "Турнир завершён. Финальный результат зафиксирован."}
	}
}
