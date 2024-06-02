RádioTaboleta
=============

This is a macOS menubar and Windows systray app for controlling the [Media Player Daemon (mpd)](https://musicpd.org/). Users can specify a list of radio streams in a config file, and playback parameters in another. The intent is to play the role of a radio receiver, now that those appliances are not as ubiquitous as before.

The mpd instance can be local or remote. One of the original goals was to be able to control an office's PA from a workstation, but it will work the same with mpd running in the workstation itself.

This app is not especially well-written or structured, it grew out of curiosity for the libraries used, the UX was developed as it went along, and it ended up meeting a demand nicely. Do not take it as a model on how to program in go.

<img align="right" width="36%" alt="macOS: menubar" src="https://github.com/entonio/RadioTaboleta/assets/5048472/5517c1d7-4e40-4659-9ee4-a18225fcc7c1"/>

On macOS, the app runs in the menubar, requiring the space usually taken by 5-6 icons in order to show the name of the station being played. Changing that is trivial if you want to, but will not provide the optimal UX.

On Windows, the app is a regular systray app, responding to right clicks.

Clicking the server name will run a shell command to restart the server. This may be useful if mpd doesn't recover properly from sleep mode, or hangs in some other way.

Clicking the song name will save it to an SQLite database file if specified in the configuration. A value `$HOME` in the path will be replaced by the user's home directory.

The configuration Stations field can be used to specify a link to download the radio stations file from. If it doesn't work for some reason, the local file is used.

The word _taboleta_ vaguely refers to the informal config file format used, and it's written precisely as intended.

Settings.taboleta
-----------------
Sample ettings file:
```
# look for mpd at localhost:32123
Mpd         localhost:32123

# set volume to 9%
Volume      9

# play each station for 10 seconds in zapping mode
Zapping     10

# resume playing on startup
Playback    start

# dial the predefined station on startup
Radio       predefined

# delete the sequence "[onAir:<any characters>]" from song titles
Trim        \[onAir:.+\]

# language (pt, en)
Language    en

# store the clicked titles at $HOME/Musicas.sqlite
SQLite      $HOME/Musicas.sqlite

# stop mpd, the OS service manager should respawn it
Restart     killall -9 mpd

# link to download the stations list from
Stations    https://some.link/to/your/playlist
```

Building
--------
The build scripts are provided in the `/scripts` folder. These are generic build + packaging scripts for go apps, and are not part of the project per se. You have to put them somewhere in your path and then call `xgob.sh` from the root of the project (the level above `/source`).

License
-------
Except where/if otherwise specified, all the files in this app are copyright of the app contributors mentioned in the `NOTICE` file and licensed under the [Apache 2.0 License](http://www.apache.org/licenses/LICENSE-2.0).
