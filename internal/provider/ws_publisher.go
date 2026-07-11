package provider

import (
	"booking-system/internal/notification/ws"
	"booking-system/pkg/shared/constants"
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
		Event: constants.EventTypeSeatLocked,
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
		Event: constants.EventTypeSeatUnlocked,
		Data: ws.SeatUnlockPayload{
			BusID:    busID,
			SeatCode: seatCode,
		},
	}

	p.hub.Broadcast(busID, msg)
	return nil
}
