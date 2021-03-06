# StreamPush - Relay

Backend RTMP relay component for StreamPush. Parses config files generated by the StreamPush frontend app and hosts an RTMP server to push those configs. I couldn't find an RTMP server that allowed for hot reloading of configs (see nginx-rtmp), so I made one in GoLang. Makes heavy use of the nareix/joy4 libraries.

## Installation

**You should be using the streampush frontend with this in the StreamPush Docker container. It can be run standalone - see the below instructions.**

1) `go get github.com/streampush/relay`
2) Make a folder somewhere named `configs`
3) Drop some JSON formatted config files in the `configs` folder. See `config-example.json` for an example config file.
4) In the same directory where `configs` is, run `relay`
5) Relay will now be running on the default ports as defined in `settings.go`.

## Usage

Relay contains a shell that's mainly used for debugging. If you're running relay standalone, it can be quite helpful. Type `help` for a list of commands.

## API
There's also an HTTP JSON API available for retrieving statistics and other info. This is available at `:8888` by default. You can change that in `settings.go`.

### `/api/reload/`
Hot reloads the config files.

### `/api/stats/`
Returns some statistics about configured streams. This includes inbound connection status, outbound connection status (per endpoint), and inbound bitrate. The `stats` object in each endpoint doesn't work - still haven't figured that one out yet.

```json
{
    "145ddbd8-97fa-4e7f-acc7-c5211d6fcde0": {
    "id": "145ddbd8-97fa-4e7f-acc7-c5211d6fcde0",
    "name": "Restream 1",
    "endpoints": {
        "451e4604-246e-4790-9cf9-18312aaaefd0": {
            "name": "YouTube",
            "url": "rtmp://a.rtmp.youtube.com/live2/streamkey",
            "connected": true,
            "stats": {
                "txBytes": 0,
                "rxBytes": 0,
                "bitrate": 0
            }
        }
    },
    "streaming": true,
    "stats": {
        "txBytes": 0,
        "rxBytes": 0,
        "bitrate": 3747.493283469538
    }
}
```