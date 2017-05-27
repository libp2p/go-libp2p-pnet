gx:
	go get github.com/whyrusleeping/gx
	go get github.com/whyrusleeping/gx-go

deps:
	gx install --global
	gx-go rewrite

