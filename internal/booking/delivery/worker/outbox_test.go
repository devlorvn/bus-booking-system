package worker

import (
	"booking-system/pkg/postgres/model"
	"booking-system/pkg/shared/constants"
	"testing"

	"github.com/google/uuid"
)

func TestOutboxWorker_StructInit(t *testing.T) {
	worker := NewOutboxWorker(nil, nil, 100)
	if worker == nil {
		t.Fatal("expected outbox worker to be initialized")
	}
}

func TestOutboxModel_RetryFields(t *testing.T) {
	outbox := model.Outbox{
		ID:         uuid.New(),
		EventType:  constants.EventTypeBookingCreated,
		Status:     "PENDING",
		RetryCount: 0,
		LastError:  "",
	}
	if outbox.RetryCount != 0 {
		t.Errorf("expected RetryCount 0, got %d", outbox.RetryCount)
	}
}
