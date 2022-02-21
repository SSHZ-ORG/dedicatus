module github.com/SSHZ-ORG/dedicatus

go 1.15

require (
	cloud.google.com/go v0.100.2
	cloud.google.com/go/storage v1.21.0
	github.com/ChimeraCoder/anaconda v2.0.1-0.20181014153429-fba449f7b405+incompatible
	github.com/ChimeraCoder/tokenbucket v0.0.0-20131201223612-c5a927568de7 // indirect
	github.com/azr/backoff v0.0.0-20160115115103-53511d3c7330 // indirect
	github.com/dustin/go-humanize v1.0.0
	github.com/dustin/go-jsonpointer v0.0.0-20160814072949-ba0abeacc3dc // indirect
	github.com/dustin/gojson v0.0.0-20160307161227-2e71ec9dd5ad // indirect
	github.com/garyburd/go-oauth v0.0.0-20180319155456-bca2e7f09a17 // indirect
	github.com/go-telegram-bot-api/telegram-bot-api/v5 v5.5.1
	github.com/gojp/kana v0.1.1-0.20200116090339-5456a3aa55f1
	github.com/julienschmidt/httprouter v1.3.0
	github.com/qedus/nds v1.0.0
	github.com/rs/xid v1.3.0
	golang.org/x/text v0.3.7
	google.golang.org/api v0.69.0
	google.golang.org/appengine/v2 v2.0.1
	google.golang.org/protobuf v1.27.1
)

replace (
	github.com/go-telegram-bot-api/telegram-bot-api/v5 => github.com/CNA-Bld/telegram-bot-api/v5 v5.5.2-0.20220221164146-714d368886bc
	github.com/qedus/nds => github.com/SSHZ-ORG/nds v1.0.1-0.20220220041449-5427bae4887c
)
