resource "aws_eks_cluster" "eks" {
  name     = var.cluster_name
  role_arn = var.labrole_arn

  vpc_config {
    subnet_ids = var.subnet_ids
    endpoint_public_access = true 
 }

  tags = {
    Name = var.cluster_name
  }
}

resource "aws_eks_node_group" "nodes" {
  cluster_name    = aws_eks_cluster.eks.name
  node_group_name = "togglemaster-nodes"
  node_role_arn   = var.labrole_arn

  subnet_ids = var.node_subnet_ids

  scaling_config {
    desired_size = 2
    max_size     = 3
    min_size     = 1
  }

  instance_types = ["t3.medium"]

  depends_on = [
    aws_eks_cluster.eks
  ]
}
