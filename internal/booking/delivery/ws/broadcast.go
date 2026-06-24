package ws

type EventType string

const (
	EventTypeSeatLocked   EventType = "seat_locked"
	EventTypeSeatUnlocked EventType = "seat_unlocked"
	EventBookingConfirmed           = "booking_confirmed"
)

type BroadcastMessage struct {
	Event EventType `json:"event"`
	Data  any       `json:"data"`
}

type SeatLockPayload struct {
	BusID      string `json:"bus_id"`
	SeatID     string `json:"seat_id"`
	SeatCode   string `json:"seat_code"`
	TempUserID string `json:"temp_user_id"`
}

type SeatUnlockPayload struct {
	BusID    string `json:"bus_id"`
	SeatID   string `json:"seat_id"`
	SeatCode string `json:"seat_code"`
}

type BookingConfirmedPayload struct {
	BookingID string   `json:"booking_id"`
	SeatCodes []string `json:"seat_codes"`
}

type RoomBroadcast struct {
	BusID   string
	Message BroadcastMessage
}
