module github.com/bjosv/kubectl-rediscluster

go 1.14

require (
	github.com/go-redis/redis/v8 v8.0.0-beta.7
	github.com/google/go-cmp v0.5.1 // indirect
	github.com/spf13/cobra v1.0.0
	github.com/spf13/pflag v1.0.5
	golang.org/x/net v0.0.0-20200625001655-4c5254603344 // indirect
	golang.org/x/xerrors v0.0.0-20200804184101-5ec99f83aff1 // indirect
	k8s.io/api v0.18.8
	k8s.io/apimachinery v0.18.8
	k8s.io/cli-runtime v0.18.8
	k8s.io/client-go v0.18.8
)
