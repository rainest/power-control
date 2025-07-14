## Overview

Reddit anecdata surveys show a pretty strong preference for CNPG. A few mention
Zalando, not many mention CrunchyData.

They (probably, at least Zalando and CNPG) provide their own Postgres container
images, because Postgres.

We'll ultimately need to figure out how we're handling shared databases in
Kustomize if we do so.

## Options

### Percona

https://docs.percona.com/percona-operator-for-postgresql/2.0/kubectl.html

NERSC currently uses Percona's MySQL operator for managing slurmdb.

AFAICT they only provide Helm or flat manifest install options. That's not too
important for service Kustomize compatibility, as we'll probably install the
operator separately and only need the one CR manifest in the service
Kustomization.

Their Postgres offering feels overly heavyweight re batteries included
decisions. It includes PGBouncer load balancing, PGBackrest backups, and their
PMM monitoring service.

You [can't disable backups](https://perconadev.atlassian.net/browse/K8SPG-427)
and I assume the component just fails without valid configuration. You can
disable PGBouncer. You can disable PMM, but you need to provide config for it
for the CR to be valid.

### Zalando

https://github.com/zalando/postgres-operator

An operator from an online EU shoe retailer (so I assume they don't sell any
paid version). It's a wrapper around their [Python DB management tool, Patroni](https://github.com/patroni/patroni).

This is what CSM services currently use. AFAIK it's been fine for those, and
we'd benefit from not needing to reinvent the wheel as much for, say, SMD.

They include a Kustomize install:

https://github.com/zalando/postgres-operator/blob/v1.14.0/manifests/kustomization.yaml

Operator config itself defaults to a giant flat ConfigMap. They do [have a CR
option](https://github.com/zalando/postgres-operator/blob/master/docs/reference/operator_parameters.md),
but the Kustomize install does not use it.

Database instance creation is handled [via CR](https://github.com/zalando/postgres-operator/blob/v1.14.0/manifests/minimal-postgres-manifest.yaml).
For example, from the [existing SMD install](https://github.com/Cray-HPE/hms-smd-charts/blob/cray-hms-smd-7.2.6/charts/v7.2/cray-hms-smd/values.yaml#L164-L186)
via the [cray-postgresql](https://github.com/Cray-HPE/base-charts/tree/master/kubernetes/cray-postgresql)
chart:

```
apiVersion: acid.zalan.do/v1
kind: postgresql
metadata:
  name: cray-smd-postgres
  namespace: services
spec:
  databases:
    hmsds: hmsdsuser
    service_db: service_account
  enableLogicalBackup: true
  logicalBackupSchedule: 10 0 * * *
  numberOfInstances: 3
  patroni:
    pg_hba:
    - local   all             all                             trust
    - local   replication     standby                         trust
    - hostssl all             +zalandos    127.0.0.1/32       pam
    - host    all             all          127.0.0.1/32       md5
    - host    all             all          127.0.0.6/32       md5
    - hostssl all             +zalandos    ::1/128            pam
    - host    all             all          ::1/128            md5
    - hostssl replication     standby      all                md5
    - hostnossl all           all          all                reject
    - hostssl all             +zalandos    all                pam
    - hostssl all             all          all                md5
  podPriorityClassName: csm-high-priority-service
  postgresql:
    version: "14"
  resources:
    limits:
      cpu: "32"
      memory: 32Gi
    requests:
      cpu: "1"
      memory: 8Gi
  teamId: cray-smd
  tls:
    secretName: cray-smd-postgres-tls
  users:
    hmsdsuser: []
    service_account: []
  volume:
    size: 100Gi
status:
  PostgresClusterStatus: Running
```

IDK exactly what all the login configuration is doing, it's just static in the
chart templates.

I do not believe there's an official CR Kustomization.

### CloudNativePG

https://github.com/cloudnative-pg/cloudnative-pg

A Kubernetes-first implementation. This may be bad in comparison to the
Patroni-based operator--the Patroni bits _may_ be portable to non-Kubernetes
installs, but lack of wrappers is good.

The project is independent OSS, but there are several companies
[offering third-party support](https://cloudnative-pg.io/support/).

Install options are apparently flat manifest or Helm. They have some Kustomize
manifests floating around, but those look like they're internal build partials
used to render the flat manifests (though other Kustomizations can presumably
then use these if desired). They also support OLM.

Unlike Zalando, they have separate Cluster and Database CRs, which is useful
for Kustomize world: we can define databases only in service manifests and let
administrators decide how to shard the databases.

Also unlike Zalando, they cannot create tables (maybe? it seems like you can
use a similar attached ConfigMap full of SQL to handle this, but it's at the
Cluster level, not Database level), though this is arguably good, since we
should use the same init for both Kubernetes and non-Kubernetes installs.

### Crunchy Postgres

https://github.com/CrunchyData/postgres-operator/

An operator from CrunchyData.

They have both Kustomize operator install and CR install examples, though the
latter are quite basic, with a pure example static CR:
https://github.com/CrunchyData/postgres-operator-examples/blob/main/kustomize/postgres/postgres.yaml

They have [paid offerings](https://www.crunchydata.com/pricing) but they appear
to be limited to their managed IaaS deployments. IDK if they use the operator
to manage those or what--if so, community support may be limited.
