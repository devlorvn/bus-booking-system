document.addEventListener("alpine:init", () => {
  Alpine.store("bus", {
    buses: [],
    selectedBus: null,

    setBuses(buses) {
      this.buses = buses;
    },

    selectBus(bus) {
      this.selectedBus = bus;
    },

    clearSelectedBus() {
      this.selectedBus = null;
    },
  });
});