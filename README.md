# aws-usage-alerts

Real-time AWS resource usage alerts grouped on AWS tags. Alerts use AWS
[SNS](https://aws.amazon.com/sns/) for email alerts. This repository includes
a service and terraform module for deployment into a kubernetes cluster.

> [!WARNING]
> Costs are provided only as an estimate. Only the AWS account provides accurate billing information.

## Current support

- Elastic file storage (EFS)
- Elastic compute (EC2)

## ⚙️ Deployment

### dev

- Create a `deploy/dev/config.yaml` file based on [config.sample.yaml](./deploy/dev/config.sample.yaml)
- Login to the AWS CLI
- Run

```bash
make dev
```

to deploy a kubernetes cluster using [k3d](https://k3d.io/v5.7.3/), the AWS resources and the `aws-usage-alerts` service.

### Production

Deploy the [aws-usage-alerts](./deploy/module) terraform module i.e

```hcl
module "aws-usage-alerts" {
  source = "github.com/ucl-arc-tre/aws-usage-alerts/module"

  image  = "ghcr.io/ucl-arc-tre/aws-usage-alerts:0.1.0"
  config = {
    groupTagKey    = "project" # Tag key to use for grouping
    storageBackend = "configMap" # Options: {inMemory, configMap}
    adminEmails    = [ # Email addresses of administrators who will receive notifications
      "alice@example.com"
    ]
    groups         = {
      example = {  # All resources tagged with project=example
        threshold = 100 # Cost threshold in $ / month
      }
    }
  }

  providers = {
    aws        = aws
    kubernetes = kubernetes
  }
}
```

## 🏗️ Development

Contributions are very welcome! Suggested steps:

- Fork this repository and create a branch.
- Install the prerequisites: k3d, terraform, docker, make, go.
- Run `pre-commit install` to install [pre-commit](https://pre-commit.com/).
- Modify, commit, push and open a pull request against `main` for review.
