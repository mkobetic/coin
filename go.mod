module github.com/mkobetic/coin

go 1.24.0

toolchain go1.24.12

require (
	github.com/aclindsa/ofxgo v0.1.3
	github.com/piquette/finance-go v1.1.1-0.20230807033903-430a57233430
	github.com/pmezard/go-difflib v1.0.0
)

replace github.com/piquette/finance-go => github.com/psanford/finance-go v0.0.0-20250222221941-906a725c60a0

require (
	github.com/aclindsa/xml v0.0.0-20201125035057-bbd5c9ec99ac // indirect
	github.com/shopspring/decimal v1.3.1 // indirect
	github.com/stretchr/testify v1.4.0 // indirect
	golang.org/x/net v0.49.0 // indirect
	golang.org/x/text v0.33.0 // indirect
	gopkg.in/yaml.v2 v2.2.8 // indirect
)
