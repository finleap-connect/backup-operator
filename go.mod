module github.com/kubism/backup-operator

go 1.13

require (
	github.com/aws/aws-sdk-go v1.30.7
	github.com/go-logr/logr v0.1.0
	github.com/go-logr/zapr v0.1.0
	github.com/google/martian v2.1.0+incompatible
	github.com/hashicorp/consul/api v1.4.0
	github.com/jessevdk/go-flags v1.4.0 // indirect
	github.com/mongodb/mongo-tools v0.0.0-20200227185201-f8447b92a52f
	github.com/mongodb/mongo-tools-common v2.0.3+incompatible
	github.com/onsi/ginkgo v1.11.0
	github.com/onsi/gomega v1.8.1
	github.com/ory/dockertest/v3 v3.5.5
	github.com/prometheus/client_golang v1.2.0
	github.com/spf13/cobra v0.0.5
	github.com/xdg/stringprep v1.0.0 // indirect
	go.mongodb.org/mongo-driver v1.3.2
	go.uber.org/zap v1.10.0
	golang.org/x/sync v0.0.0-20190423024810-112230192c58
	gopkg.in/mgo.v2 v2.0.0-20190816093944-a6b53ec6cb22 // indirect
	k8s.io/api v0.17.2
	k8s.io/apimachinery v0.17.2
	k8s.io/client-go v0.17.2
	sigs.k8s.io/controller-runtime v0.5.0
)
