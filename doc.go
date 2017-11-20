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
// Listen: tcp listening address, like ":9999", "localhost:9999", or
// "10.2.3.4:9999".
//
// LogFormat: "json" or "text".
//
// ControlSocket: unix domain socket path, like
// "/var/run/hwy3.socket". Permissions will be 0777: any local user
// can inject data to any stream. Control access by choosing a
// directory that's not world-accessible.
//
// Channels
//
// Each channel has a unique name. If the name starts with "/", the
// channel can be retrieved via HTTP using the name as the URL
// path. Otherwise, it is a private channel, useful as an input to
// other channels.
//
// Channel.name.Input: Use the output of another channel as the input
// stream. (If no Input is given, the stream can be injected by piping
// it to "hwy3 -inject name".)
//
// Channel.name.Command: pass the input stream through a shell
// command. The command is restarted automatically if it closes stdout
// or exits.
//
// Channel.name.Calm: minimum number of seconds between successive
// command restarts. Decimals are OK. Must be greater than zero;
// otherwise, defaults to 1.
//
// Channel.name.Chunk: ensure the channel outputs chunks of the given
// size (in bytes). This maintains frame sync for formats like PCM
// that have a fixed frame size.
//
// Channel.name.MP3: ensure the channel outputs whole MP3 frames. This
// maintains frame sync, but it doesn't guarantee a clean stream
// because it doesn't account for the bit reservoir.
//
// Channel.name.Buffers: maximum number of frames/chunks to buffer for
// each listener. When a listener is slow enough to fill all buffers,
// all buffered frames are dropped and the client resumes with the
// current frame.
//
// License
//
// AGPLv3.
package main
