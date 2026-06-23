document.addEventListener("alpine:init", () => {
  Alpine.data("app", () => ({
    currentPage: "bus-management",

    changePage(page) {
      this.currentPage = page;
    },
  }));
});