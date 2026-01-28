all:
	find . -type f -name '*.html' | xargs -r brotli -f -k || true
	find . -type f -name '*.html' | xargs -r gzip -f -k
	find . -type f -name '*.js' | xargs -r brotli -f -k || true
	find . -type f -name '*.js' | xargs -r gzip -f -k --best
	find . -type f -name '*.css' | xargs -r brotli -f -k || true
	find . -type f -name '*.css' | xargs -r gzip -f -k

clean:
	find . -name '*.gz' -exec rm {} \;
	find . -name '*.br' -exec rm {} \;

serve:
	go run server.go
