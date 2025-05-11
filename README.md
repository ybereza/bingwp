# BingWP - Bing Wallpaper Downloader

A simple command-line tool to download daily wallpapers from Bing. This tool automatically fetches the latest Bing wallpaper of the day and can optionally resize it to your preferred dimensions.

I personaly use this tool to set background wallpaper in [sway](https://github.com/swaywm/sway) using simple script
```bash
#!/bin/bash
set -e

bingwp -s 3840x2160 -o ~/Pictures/wallpaper.jpg
swaymsg "output * bg ~/Pictures/wallpaper.jpg center"
```

## Features

- Downloads the latest Bing wallpaper of the day
- Optional image resizing to specified dimensions
- Automatic backup of previous wallpaper
- Proxy support for restricted networks
- Simple command-line interface

## Installation

```bash
go install github.com/ybereza/bingwp@latest
```

## Usage

```bash
bingwp -s <image size> -o <output filename>
```

### Arguments

- `-o` (required): Output filename (e.g., `image.jpg` or `~/pictures/image.jpg`)
- `-s` (optional): Image size in format `WIDTHxHEIGHT` (e.g., `1920x1080`)
- `--proxy` (optional): Proxy address if needed

### Examples

Download wallpaper in original size:
```bash
bingwp -o ~/Pictures/bing-wallpaper.jpg
```

Download and resize wallpaper to 1920x1080:
```bash
bingwp -s 1920x1080 -o ~/Pictures/bing-wallpaper.jpg
```

Use with proxy:
```bash
bingwp -s 1920x1080 -o ~/Pictures/bing-wallpaper.jpg --proxy http://proxy.example.com:8080
```

## Requirements

- Go 1.24 or later
- Internet connection
- Write permissions in the output directory
