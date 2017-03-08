start: build
	./app

build:
	go generate
	go build

bundle: build
	gallium-bundle app --name Toggl2Slack --icon myapp.iconset

iconset:
	convert icon.png -resize 16x16 myapp.iconset/icon_16x16.png
	convert icon.png -resize 32x32 myapp.iconset/icon_16x16@2x.png
	convert icon.png -resize 32x32 myapp.iconset/icon_32x32.png
	convert icon.png -resize 64x64 myapp.iconset/icon_32x32@2x.png
	convert icon.png -resize 128x128 myapp.iconset/icon_128x128.png
	convert icon.png -resize 256x256 myapp.iconset/icon_128x128@2x.png
	convert icon.png -resize 256x256 myapp.iconset/icon_256x256.png
	convert icon.png -resize 512x512 myapp.iconset/icon_256x256@2x.png
	convert icon.png -resize 512x512 myapp.iconset/icon_512x512.png
	convert icon.png -resize 1024x1024 myapp.iconset/icon_512x512@2x.png
