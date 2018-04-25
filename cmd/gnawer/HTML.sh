go build -tags purego
./gnawer add -u "https://www.rbc.ru/" -n RBK -l "a.item__link" -t "div.article__header__title" -c "div.article__text"
./gnawer update