document.addEventListener("alpine:init", () => {
  Alpine.store("booking", {
    step: 1,
    seats: [],
    loadingSeats: false,
    selectedSeats: [],
    bookingResult: null,
    tempUserId: localStorage.getItem("temp_user_id") || (() => {
      const uuid = self.crypto?.randomUUID ? self.crypto.randomUUID() : 'xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx'.replace(/[xy]/g, function(c) {
        var r = Math.random() * 16 | 0, v = c == 'x' ? r : (r & 0x3 | 0x8);
        return v.toString(16);
      });
      localStorage.setItem("temp_user_id", uuid);
      return uuid;
    })(),

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