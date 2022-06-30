# Backup Operator
![Coverage](https://img.shields.io/badge/Coverage-0-red)

[![Build status](https://github.com/finleap-connect/backup-operator/actions/workflows/golang.yaml/badge.svg)](https://github.com/finleap-connect/backup-operator/actions/workflows/golang.yaml)
[![Coverage Status](https://coveralls.io/repos/github/finleap-connect/backup-operator/badge.svg?branch=main)](https://coveralls.io/github/finleap-connect/backup-operator?branch=main)
[![Go Report Card](https://goreportcard.com/badge/github.com/finleap-connect/backup-operator)](https://goreportcard.com/report/github.com/finleap-connect/backup-operator)
[![Go Reference](https://pkg.go.dev/badge/github.com/finleap-connect/backup-operator.svg)](https://pkg.go.dev/github.com/finleap-connect/backup-operator)
[![GitHub release](https://img.shields.io/github/release/finleap-connect/backup-operator.svg)](https://github.com/finleap-connect/backup-operator/releases)

## Usage

## Quick start

Add the helm repository to your list of repos:

```bash
helm repo add finleap-connect https://finleap-connect.github.io/charts/
helm repo update
```

Execute the following to get the complete list of values available:

```bash
helm show values finleap-connect/backup-operator --version <VERSION>
```

Install operator with the following command:

```bash
helm install finleap-connect/backup-operator --name myrealease --version <VERSION> --values values.yaml
```

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
For this purpose we developed a [backup operator](https://github.com/finleap-connect/backup-operator),
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

All required tools for development are automatically downloaded and stored in the `bin` sub-directory (see relevant section of [`Makefile`](./Makefile) for details).

### Testing

The tests depend on `docker` and use [`ginkgo`](https://github.com/onsi/ginkgo) and [`gomega`](https://github.com/onsi/gomega). To spin up containers for tests [`ory/dockertest`](https://github.com/ory/dockertest) is used.

#### Adding a new backup type

If you've extended the operator you need to test that the controller reconciles your new backup plan correctly. To do this, you have to add your new api type to variable `planTypes` in the file [backupplan_controller_test.go](pkg/controllers/backupplan_controller_test.go). Additionally you have to provide a function to create a new instance of your new type and add it to the variable `createTypeFuncs` in the same file. After this all controller related functionally will be tested with your newly created type as well.
