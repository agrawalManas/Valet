CREATE TABLE IF NOT EXISTS ride_reviews (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    reviewer_id UUID NOT NULL REFERENCES users(id),
    rating NUMERIC(2, 1) NOT NULL CHECK (rating >= 0 AND rating <= 5),
    review TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS rides (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    booking_id        UUID NOT NULL UNIQUE REFERENCES bookings(id),   -- one ride per booking
    site_id           UUID REFERENCES sites(id),
    ride_status       VARCHAR(20) NOT NULL DEFAULT 'REQUESTED',
    created_by        UUID NOT NULL REFERENCES users(id),
 
    pick_driver_id    UUID REFERENCES drivers(id),
    drop_driver_id    UUID REFERENCES drivers(id),
    supervisor_id     UUID REFERENCES users(id),                      -- only updated when the driver returns key to valet desk
    pick_review_id    UUID REFERENCES driver_reviews(id),
    drop_review_id    UUID REFERENCES driver_reviews(id),
    ride_review_id    UUID REFERENCES ride_reviews(id),
 
    parking_location  VARCHAR(50),                                    -- for example "B2 / slot 14"
    parked_car_images TEXT[],

    otp SMALLINT NOT NULL,
 
    created_at        TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
 
    CONSTRAINT chk_rides_status CHECK (ride_status IN (
        'REQUESTED', 'PICK_ASSIGNED', 'PICK_COMPLETED', 'PARKED', 'GMC_REQUESTED', 'DROP_ASSIGNED',
        'DROP_ENROUTE', 'DROP_COMPLETED', 'KEY_RETURNED', 'RIDE_FEEDBACK', 'COMPLETED', 'CANCELLED'))
);