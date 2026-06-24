const BusAPI = {
  async list() {
    const response = await fetch(`${API_BASE}/buses`);

    if (!response.ok) {
      throw new Error("Fetch buses failed");
    }

    const result = await response.json();
    return result.data;
  },

  async create(payload) {
    const response = await fetch(`${API_BASE}/buses`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify(payload),
    });

    if (!response.ok) {
      const err = await response.json();
      throw new Error(err.message || "Create bus failed");
    }

    return response.json();
  },
};

window.BusAPI = BusAPI;