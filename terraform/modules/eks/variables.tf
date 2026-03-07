variable "cluster_name" {
  description = "Nome do cluster"
  type        = string
  default     = "eks-togglemaster"
}

variable "labrole_arn" {
  description = "LabRole do AWS Academy"
  type        = string
  default     = "arn:aws:iam::837177764356:role/LabRole"
}

variable "vpc_id" {
  description = "ID da VPC"
  type        = string
}

variable "subnet_ids" {
  description = "Subnets do cluster"
  type        = list(string)
}

variable "node_subnet_ids" {
  description = "Subnets do nodegroup"
  type        = list(string)
}
