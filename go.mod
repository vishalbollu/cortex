// to update all:
//   rm go.mod go.sum
//   go mod tidy
//   go get -u github.com/cortexlabs/yaml@v2.2.2
//   go get -u k8s.io/client-go@v11.0.0
//   (go to the commit for the client-go release and browse to Godeps/Godeps.json to find the SHAs for k8s.io/api and k8s.io/apimachinery)
//   (replace the version for k8s.io/api with 40a48860b5abbba9aa891b02b32da429b08d96a0 and k8s.io/apimachinery with d7deff9243b165ee192f5551710ea4285dcfd615)
//   go mod tidy
//   check the diff in this file

module github.com/cortexlabs/cortex

go 1.12

require (
	github.com/aws/aws-sdk-go v1.20.1
	github.com/cortexlabs/yaml v0.0.0-20190530233410-11baebde6c89
	github.com/davecgh/go-spew v1.1.1
	github.com/docker/distribution v2.7.1+incompatible // indirect
	github.com/docker/docker v1.13.1
	github.com/docker/go-connections v0.4.0
	github.com/docker/go-units v0.4.0 // indirect
	github.com/gogo/protobuf v1.2.1 // indirect
	github.com/google/gofuzz v1.0.0 // indirect
	github.com/googleapis/gnostic v0.3.0 // indirect
	github.com/imdario/mergo v0.3.7 // indirect
	github.com/json-iterator/go v1.1.6 // indirect
	github.com/kr/pretty v0.1.0 // indirect
	github.com/mitchellh/go-homedir v1.1.0
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.1 // indirect
	github.com/opencontainers/go-digest v1.0.0-rc1 // indirect
	github.com/pkg/errors v0.8.1
	github.com/spf13/cobra v0.0.5
	github.com/stretchr/testify v1.3.0
	github.com/tcnksm/go-input v0.0.0-20180404061846-548a7d7a8ee8
	github.com/ugorji/go/codec v1.1.5-pre
	github.com/xlab/treeprint v0.0.0-20181112141820-a009c3971eca
	golang.org/x/oauth2 v0.0.0-20190604053449-0f29369cfe45 // indirect
	golang.org/x/time v0.0.0-20190308202827-9d24e82272b4 // indirect
	gopkg.in/check.v1 v1.0.0-20180628173108-788fd7840127 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	k8s.io/api v0.0.0-20190313235455-40a48860b5ab
	k8s.io/apimachinery v0.0.0-20190313205120-d7deff9243b1
	k8s.io/client-go v11.0.0+incompatible
	k8s.io/utils v0.0.0-20190607212802-c55fbcfc754a // indirect
	sigs.k8s.io/yaml v1.1.0 // indirect
)
