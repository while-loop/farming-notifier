AWS_PROFILE=default
go_apps = bin/patches bin/users bin/dynamo-trigger

bin/% : functions/%.go functions/db.go functions/models.go functions/twilio.go
	GOOS=linux go build -ldflags="-s -w" -o $@ $< functions/db.go functions/models.go functions/twilio.go

build: $(go_apps)

deploy: build
	serverless deploy

clean:
	rm -rf bin/
