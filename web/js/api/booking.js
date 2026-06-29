const BookingAPI = {
  async getSeats(busId) {
    const response = await fetch(
      `${API_BASE}/buses/${busId}/seats`
    );

    if (!response.ok) {
      throw new Error("Failed to fetch seats");
    }

    const result = await response.json();
    return result.data;
  },

  async lockSeats(busId, tempUserId, seatCodes) {
    const response = await fetch(`${API_BASE}/bookings/lock`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify({
        bus_id: busId,
        temp_user_id: tempUserId,
        seat_codes: seatCodes,
      }),
    });

    if (!response.ok) {
      const err = await response.json();
      throw new Error(err.message || "Failed to lock seats");
    }

    return response.json();
  },

  async confirmBooking(busId, tempUserId, seatCodes, name, phoneNumber, email) {
    const response = await fetch(`${API_BASE}/bookings/confirm`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify({
        bus_id: busId,
        temp_user_id: tempUserId,
        seat_codes: seatCodes,
        name: name,
        phone_number: phoneNumber,
        email: email,
      }),
    });

    if (!response.ok) {
      const err = await response.json();
      throw new Error(err.message || "Failed to confirm booking");
    }

    return response.json();
  },
};

window.BookingAPI = BookingAPI;