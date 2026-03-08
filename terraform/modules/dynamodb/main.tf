resource "aws_dynamodb_table" "analytics" {

  name         = "togglemasteranalytics"
  billing_mode = "PAY_PER_REQUEST"

  hash_key = "eventId"

  attribute {
    name = "eventId"
    type = "S"
  }

  tags = {
    Project = "togglemaster"
  }
}
