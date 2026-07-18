const protocol = window.location.protocol === "https:" ? "wss:" : "ws:";
const WS_BASE = window.location.protocol === "file:" 
  ? "ws://localhost:8082" 
  : `${protocol}//${window.location.host}`;

window.WebSocketService = function () {
  return {
    socket: null,

    connect(busId, handlers = {}) {
      if (!busId) {
        console.error("busId required");
        return;
      }

      const url = `${WS_BASE}/ws/buses/${busId}`;
      this.socket = new WebSocket(url);

      this.socket.onopen = () => {
        console.log("[WS] connected:", busId);

        if (handlers.onOpen) {
          handlers.onOpen();
        }
      };

      this.socket.onmessage = (event) => {
        try {
          const data = JSON.parse(event.data);

          console.log("[WS] message:", data);

          if (handlers.onMessage) {
            handlers.onMessage(data);
          }
        } catch (err) {
          console.error("[WS] invalid message", err);
        }
      };

      this.socket.onerror = (error) => {
        console.error("[WS] error:", error);

        if (handlers.onError) {
          handlers.onError(error);
        }
      };

      this.socket.onclose = () => {
        console.log("[WS] disconnected");

        if (handlers.onClose) {
          handlers.onClose();
        }
      };
    },

    disconnect() {
      if (this.socket) {
        this.socket.close();
        this.socket = null;
      }
    },

    send(payload) {
      if (!this.socket) return;

      if (this.socket.readyState !== WebSocket.OPEN) {
        console.warn("[WS] socket not ready");
        return;
      }

      this.socket.send(JSON.stringify(payload));
    },
  };
};