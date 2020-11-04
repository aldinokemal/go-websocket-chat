const cacheName = "chatroom"
const version = "v1"

self.addEventListener('install', (evt) => {
    self.skipWaiting();
})

importScripts('https://storage.googleapis.com/workbox-cdn/releases/5.1.2/workbox-sw.js');
// This will trigger the importScripts() for workbox.strategies and its dependencies:
const {strategies} = workbox;

// self.addEventListener('fetch', (event) => {
//     const strategi = new strategies.NetworkFirst();
//     event.respondWith(strategi.handle({request: event.request}));
// });

self.addEventListener("push", evt => {
    console.log(evt)
})

self.addEventListener("sync", evt => {
    console.log(evt)
})