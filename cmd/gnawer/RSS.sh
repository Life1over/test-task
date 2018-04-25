go build -tags purego
./gnawer add -u "https://news.rambler.ru/rss/world/" -k RSS -n RamblerWorld
./gnawer update