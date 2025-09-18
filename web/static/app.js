(() => {
  const input = document.getElementById('orderId');
  const btn = document.getElementById('loadBtn');
  const status = document.getElementById('status');
  const pre = document.getElementById('orderJson');

  if (!input || !btn || !status || !pre) return;

  const fetchOrder = async (id) => {
    status.textContent = 'Загружаем...';
    pre.style.display = 'none';
    try {
      const resp = await fetch(`/api/v1/orders/${encodeURIComponent(id)}`);
      if (!resp.ok) {
        const txt = await resp.text();
        status.textContent = `Ошибка ${resp.status}: ${txt}`;
        return;
      }
      const data = await resp.json();
      pre.textContent = JSON.stringify(data, null, 2);
      pre.style.display = 'block';
      status.textContent = 'Готово';
    } catch (e) {
      status.textContent = `Ошибка: ${e}`;
    }
  };

  btn.addEventListener('click', () => {
    const id = (input.value || '').trim();
    if (!id) {
      status.textContent = 'Введите Order UID';
      pre.style.display = 'none';
      return;
    }
    fetchOrder(id);
  });
})();


