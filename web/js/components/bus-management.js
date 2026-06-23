document.addEventListener("alpine:init", () => {
  Alpine.data("busManagement", () => ({
    loading: false,
    showCreateBusModal: false,

    newBus: {
      licensePlate: "",
      from: "",
      to: "",
      departureTime: "",
      price: 0,
      totalSeats: 24,
      rowNames: "A,B,C,D",
      seatsPerRow: 6,
    },

    get buses() {
      return Alpine.store("bus").buses;
    },

    async init() {
      await this.fetchBuses();
    },

    async fetchBuses() {
      this.loading = true;

      try {
        const buses = await BusAPI.list();

        const mapped = buses.map((bus) => ({
          id: bus.id,
          licensePlate: bus.license_plate,
          route: `${bus.from_location} → ${bus.to_location}`,
          departureTime: new Date(bus.departure_time).toLocaleString(),
          price: bus.price,
          totalSeats: bus.total_seats,
          availableSeats: bus.available_seats,
          raw: bus,
        }));

        Alpine.store("bus").setBuses(mapped);
      } catch (err) {
        console.error(err);
      } finally {
        this.loading = false;
      }
    },

    async createBus() {
      try {
        const payload = {
          license_plate: this.newBus.licensePlate,
          from_location: this.newBus.from,
          to_location: this.newBus.to,
          departure_time: new Date(
            this.newBus.departureTime
          ).toISOString(),
          price: this.newBus.price,
          total_seats: this.newBus.totalSeats,
          row_name: this.newBus.rowNames
            .split(",")
            .map((x) => x.trim())
            .filter(Boolean),
          seats_per_row: this.newBus.seatsPerRow,
        };

        await BusAPI.create(payload);

        this.resetForm();
        this.showCreateBusModal = false;
        await this.fetchBuses();
      } catch (err) {
        console.error(err);
        alert(err.message);
      }
    },

    resetForm() {
      this.newBus = {
        licensePlate: "",
        from: "",
        to: "",
        departureTime: "",
        price: 0,
        totalSeats: 24,
        rowNames: "A,B,C,D",
        seatsPerRow: 6,
      };
    },

    formatMoney(value) {
      return new Intl.NumberFormat("vi-VN").format(value) + " VND";
    },
  }));
});