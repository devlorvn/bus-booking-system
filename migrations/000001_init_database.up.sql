CREATE TABLE buses (
    id UUID PRIMARY KEY,

    license_plate VARCHAR(50) NOT NULL UNIQUE,

    from_location VARCHAR(255) NOT NULL,
    to_location VARCHAR(255) NOT NULL,

    departure_time TIMESTAMP NOT NULL,

    total_seats INT NOT NULL CHECK (total_seats > 0),

    available_seats INT NOT NULL CHECK (available_seats >= 0),

    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE users (
    id UUID PRIMARY KEY,

    name VARCHAR(255) NOT NULL,

    email VARCHAR(255),

    phone_number VARCHAR(50) NOT NULL,

    last_booking_id UUID,

    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE seats (
    id UUID PRIMARY KEY,

    bus_id UUID NOT NULL,

    seat_code VARCHAR(10) NOT NULL,

    status VARCHAR(50) NOT NULL DEFAULT 'AVAILABLE',

    created_at TIMESTAMP DEFAULT NOW(),

    CONSTRAINT fk_seat_bus
        FOREIGN KEY (bus_id)
        REFERENCES buses(id)
        ON DELETE CASCADE
);

CREATE TABLE bookings (
    id UUID PRIMARY KEY,

    bus_id UUID NOT NULL,

    user_id UUID NOT NULL,

    status VARCHAR(50) NOT NULL,

    payment_status VARCHAR(50) NOT NULL,

    expired_at TIMESTAMP,

    created_at TIMESTAMP DEFAULT NOW(),

    CONSTRAINT fk_booking_bus
        FOREIGN KEY (bus_id)
        REFERENCES buses(id),

    CONSTRAINT fk_booking_user
        FOREIGN KEY (user_id)
        REFERENCES users(id)
);

CREATE TABLE booking_seats (
    booking_id UUID NOT NULL,
    seat_id UUID NOT NULL,

    created_at TIMESTAMP DEFAULT NOW(),

    PRIMARY KEY (booking_id, seat_id),

    CONSTRAINT fk_booking_seat_booking
        FOREIGN KEY (booking_id)
        REFERENCES bookings(id)
        ON DELETE CASCADE,

    CONSTRAINT fk_booking_seat_seat
        FOREIGN KEY (seat_id)
        REFERENCES seats(id)
        ON DELETE CASCADE
);

CREATE TABLE payments (
    id UUID PRIMARY KEY,

    booking_id UUID NOT NULL,

    amount NUMERIC(10,2) NOT NULL,

    status VARCHAR(50) NOT NULL,

    provider VARCHAR(50),

    transaction_code VARCHAR(255),

    paid_at TIMESTAMP,

    created_at TIMESTAMP DEFAULT NOW(),

    CONSTRAINT fk_payment_booking
        FOREIGN KEY (booking_id)
        REFERENCES bookings(id)
);

CREATE UNIQUE INDEX idx_unique_seat_per_bus
ON seats(bus_id, seat_code);

CREATE INDEX idx_bookings_user_id
ON bookings(user_id);

CREATE INDEX idx_seats_bus_id
ON seats(bus_id);

CREATE INDEX idx_booking_seats_seat_id
ON booking_seats(seat_id);