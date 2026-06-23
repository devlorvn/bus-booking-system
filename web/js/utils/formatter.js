window.formatMoney = function (value) {
  return new Intl.NumberFormat("vi-VN").format(value) + " VND";
};

