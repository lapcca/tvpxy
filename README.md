# tvpxy
tvpxy is a light-weight proxy server. It can retrieve the RTP streams and forward it to internal LAN over HTTP.

## Features

Main features:

* Accepts an HTTP request from a client and resolves the UDP address contained in the request.
* Connect to the remote RTP server with the corresponding UDP address and retrieve the video stream.
* Offer the video stream to client over HTTP.

## Logic

The client start up video stream transmission by make HTTP request with `http://Pxy-server/rtp/<remote-rtp-address>`.

Ex. by accessing `http://Pxy-server/rtp/1.2.3.4:666` tvpxy will translate it to `rtp://1.2.3.4:666`.


## Usage

First, you need to compile the server:

```bash
make build
```

If you need to build the server for OpenWrt, you can run:

```bash
make build-openwrt-amd64 
make build-openwrt-arm 
make build-openwrt-mips
```

You can find the compiled binary files in the `build/` directory.

Then, you can run the server:

```bash
./build/tvpxy --net "eth4" --port "5566" --timeout "30s"
```
