const BookingAPI = {
  async getSeats(busId) {
    const response = await fetch(
      `http://localhost:8080/api/buses/${busId}/seats`
    );

    if (!response.ok) {
      throw new Error("Failed to fetch seats");
    }

    const result = await response.json();
    return result.data;
  },
};

window.BookingAPI = BookingAPI;