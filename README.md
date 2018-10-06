# hwy3
--
hwy3 is an http server for processing and distributing audio streams.


Radio station example

With the following configuration, hwy3 records audio from a sound card,
publishes mp3 streams at http://host.example:9999/stereo (stereo 128kbps) and
.../lofi (mono 32kbps), archives the last 24 hours (1.3GB) of the stereo stream
and the last 30 days (10.3 GB) of the mono stream, and publishes the stereo
archive at https://host.example.com:9999/ui/.

    Listen: :9999
    Channels:
      /pcm:
        Command: exec arecord --device default --format cd --file-type raw
        Chunk: 16384
        Buffers: 32
        ContentType: "audio/L16; rate=44100; channels=2"
      /stereo:
        Input: /pcm
        Command: exec lame -r -m j -s 44.1 -h -b 128 - -
        MP3: true
        Buffers: 32
        MP3Dir:
          Root: /var/db/stereo-archive  # (you must create this directory)
          BitRate: 128000
          SplitOnSize: 57600000
          SplitOnSilence: 5000000000
          PurgeOnSize: 1382400000
      /lofi:
        Input: /pcm
        Command: exec lame -r -m s -a -s 44.1 -h -b 32 - -
        MP3: true
        Buffers: 32
        MP3Dir:
          Root: /var/db/lofi-archive  # (you must create this directory)
          SplitOnSize: 14400000
          SplitOnSilence: 5000000000
          PurgeOnSize: 10368000000
    Theme:
      Title: Example Radio Station


### Deployment

Save configuration file in /etc/hwy3.yaml. Install deb or rpm package from
https://github.com/tomclegg/hwy3/releases. Check `systemctl status hwy3`.

Alternatively, run `hwy3 -config /path/to/hwy3.yaml` using your preferred
service supervisor. Logs go to stderr.


### Configuration

Listen: http address and port, like ":9999", "localhost:9999", or
"10.2.3.4:9999".

ListenTLS: https address and port, like ":8443", etc.

CertFile: path to certificate chain, like
"/var/lib/acme/live/host.example.com/fullchain".

KeyFile: path to private key, like
"/var/lib/acme/live/host.example.com/privkey".

LogFormat: "json" or "text".

ControlSocket: unix domain socket path, like "/var/run/hwy3.socket". Permissions
will be 0777: any local user can inject data to any stream. Control access by
choosing a directory that's not world-accessible.


### Channels

Each channel has a unique name. If the name starts with "/", the channel can be
retrieved via HTTP using the name as the URL path. Otherwise, it is a private
channel, useful as an input to other channels.

Channels.name.Input: Use the output of another channel as the input stream. (If
no Input is given, the stream can be injected by piping it to "hwy3 -inject name
[-chunk 1024]".)

Channels.name.Command: pass the input stream through a shell command. The
command is restarted automatically if it closes stdout or exits.

Channels.name.Calm: minimum number of seconds between successive command
restarts. Decimals are OK. Must be greater than zero; otherwise, defaults to 1.

Channels.name.Chunk: ensure the channel outputs chunks of the given size (in
bytes). This maintains frame sync for formats like PCM that have a fixed frame
size.

Channels.name.MP3: ensure the channel outputs whole MP3 frames. This maintains
frame sync, but it doesn't guarantee a clean stream because it doesn't account
for the bit reservoir.

Channels.name.Buffers: maximum number of frames/chunks to buffer for each
listener. When a listener is slow enough to fill all buffers, all buffered
frames are dropped and the client resumes with the current frame.

Channels.name.BufferLow: minimum number of frames/chunks to buffer before
sending the next frame after a listener underruns its buffer.

Channels.name.MP3Dir.Root: directory to read/write mp3 files (tNNN.mp3 and
current.mp3 where NNN is a unix timestamp representing time at EOF)

Channels.name.MP3Dir.BitRate: archived data rate in bits per second. On a public
channel, this enables the archive-browsing UI, and serves archived data at
{channelname}/A-B.mp3, where A and B are start and end times formatted as
decimal UNIX timestamps. Requires MP3Dir.BitRate.

Channels.name.MP3Dir.SplitOnSilence: enable writing to mp3dir. Start a new
output file if no data has been written for the given number of nanoseconds.

Channels.name.MP3Dir.SplitOnSize: start a new output file before current.mp3
reaches the given number of bytes.

Channels.name.MP3Dir.PurgeOnSize: when starting a new output file, delete old
files to keep the total size below the given number of bytes.


### Theme

Theme.Title: Text in UI top nav.


### License

AGPLv3.
