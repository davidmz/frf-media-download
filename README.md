# FreeFeed Media Downloader

This tool downloads all the files you have uploaded to FreeFeed.

## Usage

1. Download the latest program release from the [releases page](https://github.com/davidmz/frf-media-download/releases) and unzip the archive. Select the file suitable for your operating system.

2. Generate the access token on FreeFeed using 👉 [**this link**](https://freefeed.net/settings/app-tokens/create?title=FrF%20Media%20Download&scopes=read-my-files) 👈. Paste the generated token into the _config.ini_ file.

3. Run `frf-media-download` in console/terminal.

The files will be downloaded to the "./results" folder.

After a period of time you can run the program again to download the new files. Files that have already been downloaded will not be downloaded again.

## Troubleshooting

* If MacOS doesn't let you run the executable in Terminal because it is from an "unidentified developer", you might need to do `Open with... > Terminal` from Finder first and allow execution in a popup that appears.
