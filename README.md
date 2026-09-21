# qBitSlskd

An adapter to download files from Soulseek with Lidarr

## How it works

qBitSlskd emulates a torznab and qBitTorrent instance, allowing it to control search results and downloads.

## Setup

### Docker Compose

Add the following to your docker compose
```yaml
  qbitslskd:
    build: https://github.com/EastArctica/qBitSlskd.git
    container_name: qbitslskd
    env_file:
      - .env
    ports:
      - 3074:3000
```

Add the following fields to your .env file and modify as needed

```.env
DOWNLOAD_AUDIO_ONLY=true
SLSKD_INCOMPLETE_DIR=/data/downloads-slskd/incomplete/
SLSKD_COMPLETE_DIR=/data/downloads-slskd/complete/
SLSKD_ROOT=https://slskd.example.com
QBITSLSKD_ROOT=http://qbitslskd:3000
DELETE_SEARCHES=true
```

Once the docker is set, restart your docker container


### Lidarr

* In Lidarr, click Settings -> Indexers -> Add -> Torznab
    * Set the url to the url of your qBitSlskd instance
    * Set the API key to your SLSKD api key
* In Lidarr, click Settings -> Download Clients -> Add -> qBitTorrent
    * Set the url to the url of your qBitSlskd instance
    * Set the password to your SLSKD api key (leave the username blank)

## Use

Once set up, use like any other Lidarr setup. Currently there is no automated way to import the soulseek downloads, so instead use [beets](https://github.com/beetbox/beets) to move and format music.

