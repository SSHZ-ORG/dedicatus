# mpv_gif

A simple and stupid script that calls `ffmpeg` to produce GIF / MPEG4_GIF.

## Dependencies

You need mpv and ffmpeg installed, and have `ffmpeg` in your `PATH`.

* OSX users: Install [Homebrew](https://brew.sh/), then `brew install ffmpeg mpv`.
    * Technically this should work with [IINA](https://iina.io/) too, but you have to set up keybindings manually.
* Windows users: You can install them manually, or you can use [Chocolatey](https://chocolatey.org/) and `choco install ffmpeg mpv`.
* Linux users: Finding the correct packages to use is left as an exercise to the reader.

## Installation

Simply put [mpv_gif.lua](mpv_gif.lua) into your mpv installation's `scripts` directory.

By default, this is `~/.config/mpv/scripts` for \*nix and `%APPDATA%\mpv\scripts` for Windows. (You may have to create the `scripts` folder first.)

You may also clone this repository and create a link, so you can update it easily by `git pull`.

## Usage

Open your file with `mpv`. For BD / DVD, don't open it as `bd://` or `dvd://` devices. Open the stream files directly instead. (Or you can do `mpv VIDEO_TS` for DVD or `mpv STREAM` for BD, it will treat it as a playlist.) 

During playback, press <kbd>G</kbd> to set start time, and press <kbd>Shift</kbd>+<kbd>G</kbd> to set end time. Note that start time is inclusive but end time is exclusive. With the default keybinding, you can press <kbd>,</kbd> and <kbd>.</kbd> to jump by one frame.

Press <kbd>Ctrl</kbd>+<kbd>G</kbd> to create a MPEG4_GIF for Telegram, and <kbd>Ctrl</kbd>+<kbd>Shift</kbd>+<kbd>G</kbd> to create a MP4 with audio.

The created file will by default be put in the same directory of the original file, but if that is not writable (like if you are playing back a DVD or from a readonly network mount), it gets written to your home directory (\*nix: `~`, Windows: `%USERPROFILE%`).

## Notes

MPEG4_GIF is limited to 720p, and MP4 (with audio) is limited to 1080p. This is based on Telegram's behavior. `mp4` files that are larger than this will get sent as a video instead of a GIF.

We also convert everything to `yuv420p`, because `yuv420p10le` sometimes breaks the Android client.

On Windows the file name / path cannot contain non-ASCII characters. It appears to be a difficult problem unless we release with some `.dll` files.
