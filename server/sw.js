// Service worker del Asistente IA.
// Permite "instalar" la app y que abra aunque no haya internet (muestra la
// pantalla guardada). NO toca el chat ni las imágenes externas.
const CACHE = "asistente-v1";
const SHELL = [
  "/",
  "/manifest.webmanifest",
  "/icon-192.png",
  "/icon-512.png",
  "/favicon.png",
];

self.addEventListener("install", (e) => {
  e.waitUntil(
    caches.open(CACHE).then((c) => c.addAll(SHELL)).then(() => self.skipWaiting())
  );
});

self.addEventListener("activate", (e) => {
  e.waitUntil(
    caches
      .keys()
      .then((ks) => Promise.all(ks.filter((k) => k !== CACHE).map((k) => caches.delete(k))))
      .then(() => self.clients.claim())
  );
});

self.addEventListener("fetch", (e) => {
  const req = e.request;
  if (req.method !== "GET") return; // no tocar el chat (POST a /chat/stream)
  const url = new URL(req.url);
  if (url.origin !== location.origin) return; // no tocar imágenes de Pollinations ni externos

  // El HTML va "primero la red" para que un despliegue nuevo se vea enseguida;
  // si no hay internet, usa la copia guardada.
  if (req.mode === "navigate") {
    e.respondWith(
      fetch(req)
        .then((r) => {
          const copia = r.clone();
          caches.open(CACHE).then((c) => c.put("/", copia));
          return r;
        })
        .catch(() => caches.match("/"))
    );
    return;
  }

  // Archivos estáticos del propio servidor (iconos): primero la copia guardada.
  e.respondWith(caches.match(req).then((r) => r || fetch(req)));
});
