package service

import (
	"context"
	"log"
	"math/rand"
	"time"

	"github.com/google/uuid"
)

type FakePaymentProcessor struct {
}

func NewFakePaymentProcessor() *FakePaymentProcessor {
	return &FakePaymentProcessor{}
}

func (f *FakePaymentProcessor) Process(ctx context.Context, bookingID uuid.UUID) error {
	log.Println("processing payment:", bookingID)

	select {
	case <-ctx.Done():
		return ctx.Err()

	case <-time.After(5 * time.Second):
	}

	success := rand.Intn(100) >= 20 // 80% success

	if success {
		log.Println("PAYMENT_SUCCESS:", bookingID)
	} else {
		log.Println("PAYMENT_FAILED:", bookingID)
	}

	return nil
}
