document.addEventListener("alpine:init", () => {
  Alpine.store("booking", {
    step: 1,
    seats: [],
    loadingSeats: false,
    selectedSeats: [],
    bookingResult: null,

    setSeats(seats) {
      this.seats = seats;
    },

    toggleSeat(seatCode) {
      const idx = this.selectedSeats.indexOf(seatCode);

      if (idx >= 0) {
        this.selectedSeats.splice(idx, 1);
      } else {
        this.selectedSeats.push(seatCode);
      }
    },

    reset() {
      this.step = 1;
      this.seats = [];
      this.selectedSeats = [];
      this.bookingResult = null;
    },
  });
});