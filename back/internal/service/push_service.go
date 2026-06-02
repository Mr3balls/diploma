package service

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"esports-backend/internal/entity"
	"esports-backend/internal/repository"

	webpush "github.com/SherClockHolmes/webpush-go"
)

type PushService struct {
	subs       *repository.PushSubscriptionRepository
	vapidPriv  string
	vapidPub   string
	vapidEmail string
	enabled    bool
}

func NewPushService(subs *repository.PushSubscriptionRepository, vapidPriv, vapidPub, vapidEmail string) *PushService {
	enabled := vapidPriv != "" && vapidPub != "" && vapidEmail != ""
	if !enabled {
		log.Println("push notifications disabled: VAPID keys not configured")
	}
	return &PushService{subs: subs, vapidPriv: vapidPriv, vapidPub: vapidPub, vapidEmail: vapidEmail, enabled: enabled}
}

// VAPIDPublicKey returns the base64url-encoded public key for the browser subscription.
func (s *PushService) VAPIDPublicKey() string { return s.vapidPub }

// RegisterSubscription stores a push subscription for a user.
func (s *PushService) RegisterSubscription(ctx context.Context, userID string, sub entity.PushSubscription) error {
	sub.UserID = userID
	return s.subs.Save(ctx, &sub)
}

// UnregisterSubscription removes a push subscription.
func (s *PushService) UnregisterSubscription(ctx context.Context, userID, endpoint string) error {
	return s.subs.Delete(ctx, userID, endpoint)
}

type pushPayload struct {
	Title string `json:"title"`
	Body  string `json:"body"`
}

// SendToUser sends a Web Push notification to all registered devices of a user.
// It is designed to be called from a goroutine (fire-and-forget).
func (s *PushService) SendToUser(userID, title, body string) {
	if !s.enabled {
		return
	}
	ctx := context.Background()
	subs, err := s.subs.ListByUserID(ctx, userID)
	if err != nil || len(subs) == 0 {
		return
	}
	data, _ := json.Marshal(pushPayload{Title: title, Body: body})
	for _, sub := range subs {
		s.sendOne(ctx, sub, data)
	}
}

func (s *PushService) sendOne(ctx context.Context, sub entity.PushSubscription, payload []byte) {
	resp, err := webpush.SendNotification(payload, &webpush.Subscription{
		Endpoint: sub.Endpoint,
		Keys: webpush.Keys{
			Auth:   sub.Auth,
			P256dh: sub.P256dh,
		},
	}, &webpush.Options{
		VAPIDPrivateKey: s.vapidPriv,
		VAPIDPublicKey:  s.vapidPub,
		Subscriber:      s.vapidEmail,
		TTL:             86400,
	})
	if err != nil {
		log.Printf("push send error for sub %s: %v", sub.Endpoint[:min(20, len(sub.Endpoint))], err)
		return
	}
	defer resp.Body.Close()
	// HTTP 410 Gone means the subscription is expired — clean it up.
	if resp.StatusCode == http.StatusGone {
		_ = s.subs.DeleteByEndpoint(ctx, sub.Endpoint)
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
