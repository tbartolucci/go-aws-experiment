resource "aws_dynamodb_table" "photos-users-dynamodb-table" {
  name           = "bitsbybit_prod_math_users"
  read_capacity  = 5
  write_capacity = 5
  hash_key       = "Uuid"

  attribute {
    name = "Uuid"
    type = "S"
  }

  #  ttl {
  #    attribute_name = "TimeToExist"
  #     enabled = false
  #  }

  tags {
    Name        = "dynamodb-math-users-table"
    Environment = "production"
  }
}