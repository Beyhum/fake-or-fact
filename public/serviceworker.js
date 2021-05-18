const assetsCache = "fake-or-fact-assets"
const assets = [
    "/",
    "/css/app.css",
    "/css/bootstrap.min.css",
    "/favicon.ico",

    "https://cdn.jsdelivr.net/npm/vue@2.5.17/dist/vue.min.js",
    "https://cdn.jsdelivr.net/npm/vue-resource@1.5.1",
    "https://fonts.googleapis.com/css?family=Raleway:400,700",

    "/gif/fake-colbert.gif",
    "/gif/nope-hillary.gif",
    "/gif/trump-fake.gif",
    "/gif/trump-wrong.gif",

    "/img/fake-or-fact-icon-72.png",
    "/img/fake-or-fact-icon-192.png",
    "/img/fake-or-fact-icon-256.png",
    "/img/fake-or-fact-icon-512.png",
    "/img/fake-or-fact-icon-white.png",
    "/img/fake-or-fact-icon.png"
]

self.addEventListener("install", event => {
    event.waitUntil(
        caches.open(assetsCache).then(cache => cache.addAll(assets))
    )
})

self.addEventListener('fetch', function (event) {
    event.respondWith(
        caches.match(event.request).then(function (response) {
            return response || fetch(event.request);
        })
    );
});

// refresh the cache whenever the page is loaded and online to avoid stale assets
self.addEventListener('message', (event) => {
    if (event.data == 'refresh') {
        event.waitUntil(
            // manifest.json is never cached, if we could retrieve it, we have access to the network
            fetch("/manifest.json").then(resp => {
                caches.delete(assetsCache).then(deleted => {
                    return caches.open(assetsCache).then(cache => cache.addAll(assets))
                })
            })
            
        );
    }
});