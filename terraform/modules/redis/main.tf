resource "aws_elasticache_subnet_group" "redis" {

  name       = "togglemaster-redis-subnets"
  subnet_ids = var.subnet_ids

}

resource "aws_elasticache_cluster" "redis" {

  cluster_id           = "togglemaster-eval"
  engine               = "redis"
  node_type            = "cache.t3.micro"
  num_cache_nodes      = 1
  parameter_group_name = "default.redis7"

  subnet_group_name = aws_elasticache_subnet_group.redis.name

}
