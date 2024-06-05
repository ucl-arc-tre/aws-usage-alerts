resource "null_resource" "k3d_create" {
  provisioner "local-exec" {
    command = <<EOF
bash -c '
function cluster_exists(){
    k3d cluster list | grep -q "${local.k3d_cluster_name}"
}

if ! cluster_exists; then
  k3d cluster create "${local.k3d_cluster_name}" \
    --api-port 6550 \
    --servers 1 \
    --agents 0 \
    --wait
  sleep 30
fi
k3d kubeconfig get "${local.k3d_cluster_name}" > "${local.kubeconfig_path}"
chmod 600 "${local.kubeconfig_path}"
'
EOF
  }

  triggers = {
    always = timestamp()
  }
}

resource "null_resource" "k3d_destroy" {
  provisioner "local-exec" {
    command = "k3d cluster delete ${self.triggers.cluster_name}"
    when    = destroy
  }

  triggers = {
    cluster_name = local.k3d_cluster_name
  }
}

resource "null_resource" "import_image" {
  provisioner "local-exec" {
    command = <<EOF
docker build --tag ${local.image} --target release ${path.module}/../..
k3d image import ${local.image} -c ${local.k3d_cluster_name}
EOF
  }

  triggers = {
    "Dockerfile" = md5(file("${path.module}/../../Dockerfile"))
  }

  depends_on = [null_resource.k3d_create]
}

module "aws-usage-alerts" {
  source = "../module"

  image               = local.image
  config_file_content = local.config_content
  debug_logging       = true

  depends_on = [
    null_resource.import_image,
    null_resource.k3d_destroy
  ]
}
