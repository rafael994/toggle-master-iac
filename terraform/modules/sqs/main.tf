resource "aws_sqs_queue" "analytics_queue" {

  name = "tm-analytics-queue"

  visibility_timeout_seconds = 30

  tags = {
    Project = "togglemaster"
  }

}
