package events

type ReviewEvent struct {
	EventID    string        `json:"event_id"`
	EventType  string        `json:"event_type"`
	OccurredAt int64         `json:"occurred_at"`
	Source     string        `json:"source"`
	Payload    ReviewPayload `json:"payload"`
}
type ReviewPayload struct {
	ReviewID uint `json:"review_id"`
	ShopID   uint `json:"shop_id"`
	Score    int  `json:"score"`
}

const (
	ReviewCreated = "review.created"
	ReviewLiked   = "review.liked"
	ReviewUnliked = "review.unliked"
)
