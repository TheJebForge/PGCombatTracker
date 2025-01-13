# PG Combat Tracker

Project requires Go 1.23

Build requirements for Ubuntu:
```shell
sudo apt install build-essential libwayland-dev libxkbcommon-dev libgtk-3-dev libvulkan-dev libxkbcommon-x11-dev libx11-xcb-dev 
```

Windows Build command:
```shell
GOOS=windows GOARCH=amd64 gogio -buildmode=exe -icon=pgcticon.png -arch=amd64 -target=windows -o PGCombatTracker.exe ../PGCombatTracker
```