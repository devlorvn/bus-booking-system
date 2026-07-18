window.API_BASE = window.location.origin === "null" || window.location.protocol === "file:" 
  ? "http://localhost:8080/api" 
  : `${window.location.origin}/api`;

document.addEventListener("alpine:init", () => {
  Alpine.data("app", () => ({
    currentPage: "bus-management",

    changePage(page) {
      this.currentPage = page;
    },
  }));
});
