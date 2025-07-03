module github.com/uvalib/libra-work-import

go 1.23.0

toolchain go1.24.2

require (
	github.com/uvalib/easystore/uvaeasystore v0.0.0-20250701161005-a608a5aecb54
	github.com/uvalib/libra-metadata v0.0.0-20250513131340-aa4ee04ad7d1
)

// local development
replace github.com/uvalib/easystore/uvaeasystore => ../easystore/uvaeasystore

require (
	github.com/aws/aws-sdk-go-v2 v1.36.5 // indirect
	github.com/aws/aws-sdk-go-v2/aws/protocol/eventstream v1.6.11 // indirect
	github.com/aws/aws-sdk-go-v2/config v1.29.17 // indirect
	github.com/aws/aws-sdk-go-v2/credentials v1.17.70 // indirect
	github.com/aws/aws-sdk-go-v2/feature/ec2/imds v1.16.32 // indirect
	github.com/aws/aws-sdk-go-v2/feature/s3/manager v1.17.82 // indirect
	github.com/aws/aws-sdk-go-v2/internal/configsources v1.3.36 // indirect
	github.com/aws/aws-sdk-go-v2/internal/endpoints/v2 v2.6.36 // indirect
	github.com/aws/aws-sdk-go-v2/internal/ini v1.8.3 // indirect
	github.com/aws/aws-sdk-go-v2/internal/v4a v1.3.36 // indirect
	github.com/aws/aws-sdk-go-v2/service/cloudwatchevents v1.28.7 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/accept-encoding v1.12.4 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/checksum v1.7.4 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/presigned-url v1.12.17 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/s3shared v1.18.17 // indirect
	github.com/aws/aws-sdk-go-v2/service/s3 v1.82.0 // indirect
	github.com/aws/aws-sdk-go-v2/service/sso v1.25.5 // indirect
	github.com/aws/aws-sdk-go-v2/service/ssooidc v1.30.3 // indirect
	github.com/aws/aws-sdk-go-v2/service/sts v1.34.0 // indirect
	github.com/aws/smithy-go v1.22.4 // indirect
	github.com/lib/pq v1.10.9 // indirect
	github.com/mattn/go-sqlite3 v1.14.28 // indirect
	github.com/rs/xid v1.6.0 // indirect
	github.com/uvalib/librabus-sdk/uvalibrabus v0.0.0-20250520140939-0b78bc8b863f // indirect
	golang.org/x/exp v0.0.0-20250620022241-b7579e27df2b // indirect
)
