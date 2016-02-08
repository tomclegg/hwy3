// hwy3 is an http server for distributing a stream, like an mp3 feed,
// to many clients.
//
// Clients receive whatever hwy3 receives on stdin after they connect.
//
// Clients that receive data too slowly will miss some segments of the
// stream.
//
// Minimal example
//
// Clients receive random bytes.
//
//   hwy3 </dev/urandom
//
// PCM radio station
//
// Clients receive raw PCM data. Clients that receive data too slowly
// will miss multiples of 1000 bytes, so they don't lose sync.
//
//   arecord --device default:0 --format cd --file-type raw \
//       | hwy3 -listen :44100 -chunk 1000 -mime-type "audio/L16; rate=44100; channels=2"
//
// MP3 radio station
//
// Clients receive MP3 frames. Clients that receive data too slowly
// will miss entire MP3 frames, so they don't lose sync.
//
//   curl --retry 9999999 --retry-delay 1 http://0.0.0.0:44100/ \
//       | lame -r -h -b 128 - - \
//       | hwy3 -listen :12800 -mp3
//
// Log messages
//
// hwy3 prints a log message on stderr whenever a client connects or
// disconnects.
//
//   2016/02/08 00:39:12.371135 3 +"127.0.0.1:48663"
//
// A client connected ("+") from 127.0.0.1 port 48663, for a total of
// 3 clients connected now.
//
//   2016/02/08 00:39:20.478136 2 -"127.0.0.1:48663" 8.107070208s 127831 =15767B/s ""
//
// The client at 127.0.0.1:48663 disconnected ("-"). 2 other clients
// are still connected. We sent this client 127831 bytes in the 8.1
// seconds it was connected (an average of 15767 bytes per second). No
// errors were encountered ("").
//
package main
