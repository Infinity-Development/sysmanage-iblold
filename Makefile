api:
	CGO_ENABLED=0 go build -v 
full:
	cd frontend && npm i && npm run build && cd .. && CGO_ENABLED=0 go build -v
