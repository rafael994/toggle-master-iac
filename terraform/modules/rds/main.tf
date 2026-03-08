resource "aws_db_subnet_group" "postgres" {

  name       = "togglemaster-db-subnet-group"
  subnet_ids = var.subnet_ids

}

resource "aws_db_instance" "auth" {

  identifier = "auth-db"

  engine         = "postgres"
  instance_class = "db.t3.micro"

  allocated_storage = 20

  username = "postgres"
  password = "postgres123"

  db_subnet_group_name = aws_db_subnet_group.postgres.name

  skip_final_snapshot = true

}

resource "aws_db_instance" "flags" {

  identifier = "flags-db"

  engine         = "postgres"
  instance_class = "db.t3.micro"

  allocated_storage = 20

  username = "postgres"
  password = "postgres123"

  db_subnet_group_name = aws_db_subnet_group.postgres.name

  skip_final_snapshot = true

}

resource "aws_db_instance" "targeting" {

  identifier = "targeting-db"

  engine         = "postgres"
  instance_class = "db.t3.micro"

  allocated_storage = 20

  username = "postgres"
  password = "postgres123"

  db_subnet_group_name = aws_db_subnet_group.postgres.name

  skip_final_snapshot = true

}
