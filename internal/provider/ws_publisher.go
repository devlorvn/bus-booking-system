package provider

import (
	"booking-system/internal/booking/delivery/ws"
	"context"

	"github.com/google/uuid"
)

type WSEventPublisher struct {
	hub *ws.Hub
}

func NewWSEventPublisher(hub *ws.Hub) *WSEventPublisher {
	return &WSEventPublisher{
		hub: hub,
	}
}

func (p *WSEventPublisher) PublishSeatLocked(
	busID string,
	seatID string,
	seatCode string,
	tempUserID string,
) error {
	msg := ws.BroadcastMessage{
		Event: ws.EventTypeSeatLocked,
		Data: ws.SeatLockPayload{
			BusID:      busID,
			SeatID:     seatID,
			SeatCode:   seatCode,
			TempUserID: tempUserID,
		},
	}

	p.hub.Broadcast(busID, msg)
	return nil
}

func (p *WSEventPublisher) PublishSeatReleased(
	busID string,
	seatCode string,
) error {
	msg := ws.BroadcastMessage{
		Event: ws.EventTypeSeatUnlocked,
		Data: ws.SeatUnlockPayload{
			BusID:    busID,
			SeatCode: seatCode,
		},
	}

	p.hub.Broadcast(busID, msg)
	return nil
}

func (p *WSEventPublisher) PublishPaymentRequested(
	ctx context.Context,
	bookingID uuid.UUID,
) error {
	// No-op or log if needed, as this is local websocket and not real payment queue
	return nil
}

