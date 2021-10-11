# Backup Operator

[![Go Documentation](https://img.shields.io/badge/go-doc-blue.svg?style=flat)](https://pkg.go.dev/mod/github.com/kubism/backup-operator?tab=packages)
[![Build Backup Operator](https://github.com/kubism/backup-operator/actions/workflows/backup-operator-docker.yml/badge.svg)](https://github.com/kubism/backup-operator/actions/workflows/backup-operator-docker.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/kubism/backup-operator)](https://goreportcard.com/report/github.com/kubism/backup-operator)
[![Coverage Status](https://coveralls.io/repos/github/kubism/backup-operator/badge.svg?branch=master)](https://coveralls.io/github/kubism/backup-operator?branch=master)
[![Docker Image Version (latest semver)](https://img.shields.io/docker/v/kubismio/backup-operator.svg?sort=semver)](https://hub.docker.com/r/kubismio/backup-operator/tags)
[![Maintainability](https://api.codeclimate.com/v1/badges/5f5e31a56c7c0555121a/maintainability)](https://codeclimate.com/github/kubism/backup-operator/maintainability)

## Usage

### Setup

Find the helm chart for the backup-operator at the [Kubism.io Helm Charts](https://kubism.github.io/charts/#chart-backup-operator).

### Backup for MongoDB

Let's assume you want to backup a MongoDB replicaset. The only MongoDB
specific configuration required is the [MongoDB URI](https://docs.mongodb.com/manual/reference/connection-string/).
However you'll want to insert the sensitive data using environment variables.

For example, let's assume you have two pre-existing secrets:

* secret containing the password for the MongoDB user
* secret containing the S3 credentials (and optional encryption key for [SSE feature](https://docs.aws.amazon.com/AmazonS3/latest/dev/UsingServerSideEncryption.html))

**Note:** The below YAML mixes both [kubernetes environment composition](https://kubernetes.io/docs/tasks/inject-data-application/define-environment-variable-container/#using-environment-variables-inside-of-your-config)
in the `env` section and job environment substitution in the other parts.

The you might compose a `MongoDBBackupPlan` as in [`backup_v1alpha1_mongodbbackupplan.yaml`](./config/samples/backup_v1alpha1_mongodbbackupplan.yaml).

The above specification will create a `CronJob` with the same name and the above
`env` and also create a `Secret` with the rest of the specification and mount it
into the `CronJob` as well.

### Backup for Consul

For Consul the procedure is the same as above. However instead of providing
the URI, the `ConsulBackupPlan` requires the follow fields: `address`, `username` and `password`,
which hopefully are self-explanatory.

See example configuration in [`backup_v1alpha1_consulbackupplan.yaml`](./config/samples/backup_v1alpha1_consulbackupplan.yaml).

## Design

A common procedure of any production environments are backups.
For this purpose we developed a [backup operator](https://github.com/kubism/backup-operator),
which can be used to setup a `CronJob`, which will take care of the backup for you.

The plan specification consists of several fields and an environment specification.
This duality is very important as **environment variables should be used to pass
sensitive data** to the resulting `CronJob`.

The operator will spawn a vanilla `CronJob` and setup the environment as specified
by you. Once the job runs it will use environment substitution to replace any
variables in your specification.

Therefore you should use the `valueFrom.secretKeyRef` to provide the sensitive
parts of your environment.

The backup job will also push metrics into a prometheus pushgateway, if configured.

Once a job is finished, it will make sure to remove obsolete backups as specified
by your `retention`.

## Development

### Tools

All required tools for development are automatically downloaded and stored in the `tools` sub-directory (see relevant section of [`Makefile`](./Makefile) for details).
A custom [`tools/goget-wrapper`](./tools/goget-wrapper) is used to create a temporary isolated environment to install and compile go tools.
To make sure those can be properly used in tests, several helpers were implemented in [`pkg/testutil`](./pkg/testutil) (e.g. `HelmEnv`, `KindEnv`).

### Testing

The tests depend on `docker` and `kind` and use [`ginkgo`](https://github.com/onsi/ginkgo) and [`gomega`](https://github.com/onsi/gomega). To spin up containers for tests [`ory/dockertest`](https://github.com/ory/dockertest) is used. For the controller tests `kind` is used, which has the advantage, compared to the more lightweight kubebuilder assets approach, to properly handle finalizers and allow integration tests.

#### Adding a new backup type

If you've extended the operator you need to test that the controller reconciles your new backup plan correctly. To do this, you have to add your new api type to variable `planTypes` in the file [backupplan_controller_test.go](pkg/controllers/backupplan_controller_test.go). Additionally you have to provide a function to create a new instance of your new type and add it to the variable `createTypeFuncs` in the same file. After this all controller related functionally will be tested with your newly created type as well.

### Kubebuilder

This project uses a different project layout than what is generated by
kubebuilder. The layout adheres to the [golang standards](https://github.com/golang-standards/project-layout) layout.
For this to properly work a wrapper is required ([`tools/kubebuilder-wrapper`](./tools/kubebuilder-wrapper)),
which makes sure the correct kubebuilder version is available and temporarily
moves files around as required.

While this is certainly not beautiful, this should improve with future versions
of kubebuilder and their plugin capabilities.

#### Known quirks

* When using the kubebuilder CLI to create a new API [`main.go`](./cmd/manager/main.go)
has a wrong controllers import path and has to be fixed manually afterwards.

### Extending the operator

To extend the operator you have to use the wrapper ([`tools/kubebuilder-wrapper`](./tools/kubebuilder-wrapper)) to scaffold out a new [Kind](https://book.kubebuilder.io/cronjob-tutorial/gvks.html#kinds-and-resources) and corresponding controller. The following command (see the official [kubebuilder docs](https://book.kubebuilder.io/cronjob-tutorial/new-api.html)) must be used:

```bash
./tools/kubebuilder create api --group backup --version <version> --kind <SomeBackupPlan>
```

Using the command above will generate several classes for you:

* `api/<version>/<somebackupplan>_types.go`
  * Add the spec and other necessary stuff for your new Kind here
* `controllers/<somebackupplan>_controller.go`
  * Please add the annotations for your type to the `Reconcile` function in file [backupplan_controller.go](pkg/controllers/backupplan_controller.go) and delete the generated controller file

Please have a look at the existing types, e.g. the [mongodbbackupplan_types.go](api/v1alpha1/mongodbbackupplan_types.go). All backup plans use the base types provided in [backupplan_types.go](api/v1alpha1/backupplan_types.go) for the general backup plan settings. Additional settings needed must be created by you, like it has been done for the existing plans. New backup plan types have to implement the interface `BackupPlan` so that the generic controller implementation will work for your new type.

In addition to the operator specifics you have to implement a new command as part of the worker below `cmd/worker` like for the existing ones, e.g. [`mongodb`](cmd/worker/mongodb.go).

Please add tests for all new parts added to the operator.
