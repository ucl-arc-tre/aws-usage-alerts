# aws-usage-alerts

Real-time AWS resource usage alerts grouped on AWS tags. Alerts use AWS
[SNS](https://aws.amazon.com/sns/) for email alerts. This repository includes
a Go service and terraform module for deployment into a k8s cluster.

> [!WARNING]
> Costs are provided only as an estimate. Only the AWS account provides accurate billing information.

## Current support

- Elastic file storage (EFS)
- Elastic compute compute (EC2)

## ⚙️ Deployment

### dev

- Create a `config.yaml` file based on [config.sample.yaml](./config.sample.yaml)
- Login to AWS
- Run

```bash
make dev
```

### Production

- Create a `config.yaml` file based on [config.sample.yaml](./config.sample.yaml)
- Deploy the [aws-usage-alerts](./deploy/module) terraform module i.e

```hcl
module "aws-usage-alerts" {
  source = "github.com/ucl-arc-tre/aws-usage-alerts/module"

  image  = "ghcr.io/ucl-arc-tre/aws-usage-alerts:0.1.0"
  config = yamldecode(file("config.yaml"))
}
```

## 🏗️ Development

Contributions are very welcome! Suggested steps:

- Clone this repository and create a branch.
- Install the prerequisites: {k3d, terraform, docker, make, go}
- Run `pre-commit install` to install [pre-commit](https://pre-commit.com/).
- Modify, commit, push and open a pull request against `main` for review.
