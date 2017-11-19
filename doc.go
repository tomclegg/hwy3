/*
   hwy3, an http server for distributing audio streams
   Copyright (C) 2017  Tom Clegg

   This program is free software: you can redistribute it and/or modify
   it under the terms of the GNU Affero General Public License as published by
   the Free Software Foundation, either version 3 of the License, or
   (at your option) any later version.

   This program is distributed in the hope that it will be useful,
   but WITHOUT ANY WARRANTY; without even the implied warranty of
   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
   GNU Affero General Public License for more details.

   You should have received a copy of the GNU Affero General Public License
   along with this program.  If not, see <http://www.gnu.org/licenses/>.
*/

// hwy3 is an http server for processing and distributing audio
// streams.
//
// Radio station example
//
// To record audio from a sound card and publish multiple mp3 streams
// at http://host.example:9999/high (stereo 128kbps) and .../low (mono
// 32kbps), save this in ./config.yaml, then run hwy3.
//
//   Listen: :9999
//   LogFormat: json
//   Channels:
//     /pcm:
//       Command: exec arecord --device default --format cd --file-type raw
//       Chunk: 16384
//       Buffers: 32
//       ContentType: "audio/L16; rate=44100; channels=2"
//     /high:
//       Input: /pcm
//       Command: exec lame -r -m j -s 44.1 -h -b 128 - -
//       MP3: true
//       Buffers: 32
//     /low:
//       Input: /pcm
//       Command: exec lame -r -m s -a -s 44.1 -h -b 32 - -
//       MP3: true
//       Buffers: 32
//
// Listen
//
// Listen can look like ":9999", "localhost:9999", or "10.2.3.4:9999".
//
// LogFormat
//
// LogFormat can be "json" or "text".
//
// Channels
//
// Each channel has a unique name. If the name starts with "/", the
// channel can be retrieved via HTTP using the name as the URL
// path. Otherwise, it is a private channel, useful as an input to
// other channels.
//
// Channel configuration
//
// "Command" starts a shell command and uses its output as the stream
// data. If "Input" is given, the specified stream is passed to the
// command's stdin. The command is restarted automatically if it
// closes stdout or exits.
//
// "Calm" is the minimum number of seconds between successive command
// restarts. Decimals are OK. Must be greater than zero; otherwise,
// defaults to 1.
//
// "Chunk" ensures the channel outputs the given number of bytes at a
// time. This maintains frame sync for formats with fixed frame sizes,
// like PCM.
//
// "MP3" ensures the channel outputs whole MP3 frames. This maintains
// frame sync, but it doesn't guarantee a clean stream: it doesn't
// account for the bit reservoir.
//
// "Buffers" is the maximum number of frames buffered for each
// listener. When a listener falls this far behind, all buffered
// frames are dropped and the client resumes with the current frame.
//
// License
//
// AGPLv3.
package main
