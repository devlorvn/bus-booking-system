document.addEventListener("alpine:init", () => {
  Alpine.data("bookingFlow", () => ({
    booking: false,
    ws: null,

    init() {
      this.ws = WebSocketService();
    },
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

    get tempUserId() {
      return Alpine.store("booking").tempUserId;
    },

    async selectBus(bus) {
      Alpine.store("bus").selectBus(bus);
      Alpine.store("booking").loadingSeats = true;

      this.connectBusRoom(bus.id)

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
      if (seat.status === 'BOOKED' || seat.status === 'LOCKED') {
        return;
      }
      const idx = this.selectedSeats.indexOf(seat.code);

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

    async bookSeats() {
      if (this.selectedSeats.length === 0 || this.booking) return;

      this.booking = true;
      try {
        await BookingAPI.lockSeats(
          this.selectedBus.id,
          this.tempUserId,
          this.selectedSeats
        );

        alert(`Đặt vé thành công cho các ghế: ${this.selectedSeats.join(', ')}`);

        // Reset booking flow
        Alpine.store("booking").reset();
        Alpine.store("bus").clearSelectedBus();
      } catch (err) {
        console.error(err);
        alert(`Đặt vé thất bại: ${err.message}`);
      } finally {
        this.booking = false;
      }
    },

    connectBusRoom(busId) {
      if (!this.ws) return;

      this.ws.disconnect();

      this.ws.connect(busId, {
        onOpen: () => {
          console.log("[BookingFlow] WS connected");
        },

        onMessage: (message) => {
          this.handleSocketEvent(message);
        },

        onClose: () => {
          console.log("[BookingFlow] WS closed");
        },

        onError: (err) => {
          console.error("[BookingFlow] WS error", err);
        },
      });
    },

    handleSocketEvent(message) {
      switch (message.event) {
        case "seat_locked":
          this.handleSeatLocked(message.data);
          break;

        case "seat_released":
          this.handleSeatReleased(message.data);
          break;

        default:
          console.warn("Unknown WS event:", message.event);
      }
    },
    handleSeatLocked(data) {
      const seat = this.seats.find(
        (s) => s.code === data.seat_code
      );

      if (!seat) return;

      seat.status = 'LOCKED';

      console.log("Seat locked:", seat.code);
    },
    handleSeatReleased(data) {
      const seat = this.seats.find(
        (s) => s.code === data.seat_code
      );

      if (!seat) return;

      seat.status = 'AVAILABLE';

      console.log("Seat released:", seat.code);
    },

    disconnect() {
      if (this.ws) {
        this.ws.disconnect();
      }
    },

    formatMoney(value) {
      return new Intl.NumberFormat("vi-VN").format(value) + " VND";
    },
  }));
});