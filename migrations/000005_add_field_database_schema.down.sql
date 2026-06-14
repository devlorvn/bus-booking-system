DROP INDEX IF EXISTS idx_booking_seats_seat_id;
DROP INDEX IF EXISTS idx_booking_seats_booking_id;

DROP TABLE IF EXISTS booking_seats;

ALTER TABLE users
DROP CONSTRAINT IF EXISTS users_phone_number_key;

ALTER TABLE users
ADD CONSTRAINT users_email_key UNIQUE (email);

ALTER TABLE bookings
DROP COLUMN IF EXISTS total_amount,
DROP COLUMN IF EXISTS total_seats;