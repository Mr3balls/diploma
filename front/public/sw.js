self.addEventListener('push', (event) => {
  let data = { title: 'Уведомление', body: '' };
  try { data = event.data ? event.data.json() : data; } catch (_) {}

  event.waitUntil(
    self.registration.showNotification(data.title, {
      body: data.body || '',
      icon: '/favicon.ico',
      badge: '/favicon.ico',
      data: data,
      vibrate: [200, 100, 200],
    })
  );
});

self.addEventListener('notificationclick', (event) => {
  event.notification.close();
  event.waitUntil(
    clients.matchAll({ type: 'window', includeUncontrolled: true }).then((windowClients) => {
      for (const client of windowClients) {
        if ('focus' in client) return client.focus();
      }
      if (clients.openWindow) return clients.openWindow('/notifications');
    })
  );
});
