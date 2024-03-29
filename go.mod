module github.com/SSHZ-ORG/dedicatus

go 1.19

require (
	cloud.google.com/go/storage v1.29.0
	github.com/go-telegram-bot-api/telegram-bot-api/v5 v5.5.1
	github.com/gojp/kana v0.1.1-0.20200116090339-5456a3aa55f1
	github.com/julienschmidt/httprouter v1.3.0
	github.com/qedus/nds v1.0.0
	github.com/rs/xid v1.4.0
	golang.org/x/text v0.8.0
	google.golang.org/api v0.112.0
	google.golang.org/appengine/v2 v2.0.2
	google.golang.org/protobuf v1.29.0
)

require (
	cloud.google.com/go v0.110.0 // indirect
	cloud.google.com/go/compute v1.18.0 // indirect
	cloud.google.com/go/compute/metadata v0.2.3 // indirect
	cloud.google.com/go/iam v0.12.0 // indirect
	github.com/golang/groupcache v0.0.0-20200121045136-8c9f03a8e57e // indirect
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/google/go-cmp v0.5.9 // indirect
	github.com/google/uuid v1.3.0 // indirect
	github.com/googleapis/enterprise-certificate-proxy v0.2.3 // indirect
	github.com/googleapis/gax-go/v2 v2.7.0 // indirect
	go.opencensus.io v0.24.0 // indirect
	golang.org/x/net v0.8.0 // indirect
	golang.org/x/oauth2 v0.6.0 // indirect
	golang.org/x/sys v0.6.0 // indirect
	golang.org/x/xerrors v0.0.0-20220907171357-04be3eba64a2 // indirect
	google.golang.org/appengine v1.6.7 // indirect
	google.golang.org/genproto v0.0.0-20230303212802-e74f57abe488 // indirect
	google.golang.org/grpc v1.53.0 // indirect
)

replace (
	github.com/go-telegram-bot-api/telegram-bot-api/v5 => github.com/CNA-Bld/telegram-bot-api/v5 v5.5.2-0.20220221164146-714d368886bc
	github.com/qedus/nds => github.com/SSHZ-ORG/nds v1.0.1-0.20220220041449-5427bae4887c
)
