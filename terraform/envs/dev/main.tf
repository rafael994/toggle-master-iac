module "networking" {
  source = "../../modules/networking"

  vpc_cidr            = "10.0.0.0/16"
  public_subnet_cidr  = "10.0.1.0/24"
  private_subnet_cidr = "10.0.2.0/24"
  region              = "us-east-1"
}

module "eks" {
  source = "../../modules/eks"

  vpc_id = module.networking.vpc_id

  subnet_ids = module.networking.public_subnets

  node_subnet_ids = module.networking.public_subnets
}
module "ecr" {

  source = "../../modules/ecr"

  repositories = [
    "auth-service",
    "flags-service",
    "targeting-service",
    "evaluation-service",
    "analytics-service"
  ]

}
module "dynamodb" {
  source = "../../modules/dynamodb"
}

module "sqs" {
  source = "../../modules/sqs"
}

module "redis" {

  source = "../../modules/redis"

  subnet_ids = module.networking.private_subnets

}

module "rds" {

  source = "../../modules/rds"

  subnet_ids = module.networking.private_subnets

}
