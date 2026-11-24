// Service worker de Sillage. Il ne met rien en cache : le frontend est
// embarqué dans le binaire et revalidé par ETag, une copie locale ne ferait que
// servir un app.js périmé après une mise à jour.
//
// Son seul rôle : quand la fenêtre installée s'ouvre alors que le serveur ne
// tourne pas, afficher une page qui le dit (et se recharge seule dès qu'il
// répond) au lieu de la page d'erreur du navigateur.
//
// Seules les navigations passent par lui. L'API, le flux SSE, les ressources
// statiques et les recettes vont directement au réseau : la règle de routage
// statique (Chrome 123 et plus) les sort du service worker avant même qu'il
// démarre, et le filtre du gestionnaire `fetch` couvre les autres navigateurs.

self.addEventListener('install', function (event) {
  self.skipWaiting();
  if (event.addRoutes) {
    event.waitUntil(event.addRoutes([
      { condition: { not: { requestMode: 'navigate' } }, source: 'network' },
      { condition: { urlPattern: { pathname: '/preview/*' } }, source: 'network' },
      { condition: { urlPattern: { pathname: '/api/*' } }, source: 'network' }
    ]).catch(function () { /* règle refusée : le filtre plus bas suffit */ }));
  }
});

self.addEventListener('activate', function (event) {
  event.waitUntil(Promise.all([
    self.clients.claim(),
    // La navigation part vers le réseau pendant que le service worker démarre,
    // au lieu d'attendre son réveil.
    self.registration.navigationPreload ? self.registration.navigationPreload.enable() : null
  ]));
});

self.addEventListener('fetch', function (event) {
  var req = event.request;
  if (req.mode !== 'navigate') return;
  var path = new URL(req.url).pathname;
  if (path.indexOf('/preview/') === 0 || path.indexOf('/api/') === 0) return;
  event.respondWith(
    Promise.resolve(event.preloadResponse)
      .then(function (preloaded) { return preloaded || fetch(req); })
      .catch(offlinePage)
  );
});

function offlinePage() {
  var fr = (self.navigator.language || '').toLowerCase().indexOf('fr') === 0;
  var title = fr ? 'Sillage ne tourne pas' : 'Sillage is not running';
  var body = fr
    ? 'Lancez <code>sillage</code> dans un terminal, ou <code>brew services start sillage</code> pour qu’il démarre avec votre session. Cette fenêtre se recharge d’elle-même dès qu’il répond.'
    : 'Start <code>sillage</code> in a terminal, or <code>brew services start sillage</code> to have it start with your session. This window reloads by itself as soon as it answers.';
  var html = '<!doctype html><html lang="' + (fr ? 'fr' : 'en') + '"><head><meta charset="utf-8">' +
    '<meta name="viewport" content="width=device-width, initial-scale=1"><title>Sillage</title>' +
    '<link rel="icon" href="/favicon.svg" type="image/svg+xml"><style>' +
    'body{margin:0;min-height:100vh;display:grid;place-items:center;background:#f2f0eb;color:#191917;' +
    'font:15px/1.55 -apple-system,BlinkMacSystemFont,"SF Pro Text",Inter,"Segoe UI",system-ui,sans-serif}' +
    'main{max-width:440px;padding:32px 24px}h1{font-size:18px;font-weight:600;margin:0 0 8px}' +
    'p{margin:0;color:#5b5a55}code{font:13px ui-monospace,SFMono-Regular,Menlo,monospace;' +
    'background:rgba(25,25,23,.06);padding:1px 5px;border-radius:4px}</style></head><body><main>' +
    '<h1>' + title + '</h1><p>' + body + '</p></main><script>' +
    'setInterval(function(){fetch("/",{cache:"no-store"}).then(function(r){if(r.ok)location.reload()},function(){})},2000)' +
    '</script></body></html>';
  return new Response(html, { status: 503, headers: { 'Content-Type': 'text/html; charset=utf-8', 'Cache-Control': 'no-store' } });
}
