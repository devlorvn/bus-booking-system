document.addEventListener("alpine:init", () => {
  Alpine.data("bookingFlow", () => ({
    get buses() {
      return Alpine.store("bus").buses;
    },

    get selectedBus() {
      return Alpine.store("bus").selectedBus;
    },

    get seats() {
      return Alpine.store("booking").seats;
    },

    get selectedSeats() {
      return Alpine.store("booking").selectedSeats;
    },

    get step() {
      return Alpine.store("booking").step;
    },

    async selectBus(bus) {
      Alpine.store("bus").selectBus(bus);
      Alpine.store("booking").loadingSeats = true;

      try {
        const seats = await BookingAPI.getSeats(bus.id);

        Alpine.store("booking").setSeats(
          seats.map((seat) => ({
            id: seat.id,
            code: seat.seat_code,
            status: seat.status,
          }))
        );

        Alpine.store("booking").step = 2;
      } catch (err) {
        console.error(err);
      } finally {
        Alpine.store("booking").loadingSeats = false;
      }
    },

    toggleSeat(seat) {
      if (seat.status === "BOOKED") return;
      Alpine.store("booking").toggleSeat(seat.code);
    },

    isSelected(code) {
      return this.selectedSeats.includes(code);
    },

    totalAmount() {
      if (!this.selectedBus) return 0;
      return this.selectedSeats.length * this.selectedBus.price;
    },

    goBack() {
      Alpine.store("booking").reset();
      Alpine.store("bus").clearSelectedBus();
    },

    formatMoney(value) {
      return new Intl.NumberFormat("vi-VN").format(value) + " VND";
    },
  }));
});