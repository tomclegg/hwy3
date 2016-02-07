// hwy3 is an http server for distributing a stream, like an mp3 feed,
// to many clients.
//
// Clients receive whatever hwy3 receives on stdin after they connect.
//
// Clients that receive data too slowly will miss some segments of the
// stream.
//
// Short version:
//
//   hwy3 </dev/urandom
//
// Slightly more realistic:
//
//   curl --retry 9999999 --retry-delay 1 http://source.example/ | hwy3 -listen :8000
//
// TODO
//
// -mp3: send clients only whole MP3 frames
//
package main
