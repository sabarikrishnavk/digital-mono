module github.com/omni-compos/digital-mono/services/user

go 1.21

require (
	github.com/google/uuid v1.6.0
	github.com/gorilla/mux v1.8.1
	github.com/graphql-go/graphql v0.8.1
	github.com/graphql-go/handler v0.2.4
	github.com/lib/pq v1.10.9
	github.com/omni-compos/digital-mono/libs/database v0.0.0
	github.com/omni-compos/digital-mono/libs/logger v0.0.0
	github.com/omni-compos/digital-mono/libs/metrics v0.0.0
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/stretchr/objx v0.5.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace (
	github.com/omni-compos/digital-mono/libs/auth => ../../libs/auth
	github.com/omni-compos/digital-mono/libs/database => ../../libs/database
	github.com/omni-compos/digital-mono/libs/logger => ../../libs/logger
	github.com/omni-compos/digital-mono/libs/metrics => ../../libs/metrics
)

require github.com/stretchr/testify v1.8.4 // indirect , for testing
