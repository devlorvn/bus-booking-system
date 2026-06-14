-- =========================
-- BOOKINGS
-- =========================

ALTER TABLE bookings
ADD COLUMN total_amount NUMERIC(10,2) NOT NULL DEFAULT 0,
ADD COLUMN total_seats INT NOT NULL DEFAULT 0;

-- =========================
-- USERS
-- =========================

-- Drop old email unique constraint
ALTER TABLE users
DROP CONSTRAINT IF EXISTS users_email_key;

-- Add phone unique constraint
ALTER TABLE users
ADD CONSTRAINT users_phone_number_key UNIQUE (phone_number);

-- =========================
-- BOOKING_SEATS
-- =========================

CREATE TABLE booking_seats (
    booking_id UUID NOT NULL,
    seat_id UUID NOT NULL,

    PRIMARY KEY (booking_id, seat_id),

    CONSTRAINT fk_booking_seats_booking
        FOREIGN KEY (booking_id)
        REFERENCES bookings(id)
        ON DELETE CASCADE,

    CONSTRAINT fk_booking_seats_seat
        FOREIGN KEY (seat_id)
        REFERENCES seats(id)
        ON DELETE CASCADE
);

CREATE INDEX idx_booking_seats_booking_id
ON booking_seats(booking_id);

CREATE INDEX idx_booking_seats_seat_id
ON booking_seats(seat_id);