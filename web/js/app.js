window.API_BASE = "http://localhost:8080/api";

document.addEventListener("alpine:init", () => {
  Alpine.data("app", () => ({
    currentPage: "bus-management",

    changePage(page) {
      this.currentPage = page;
    },
  }));
});
