ALTER TABLE buses
ADD CONSTRAINT buses_license_plate_key
UNIQUE (license_plate);