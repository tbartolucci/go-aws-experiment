resource "aws_dynamodb_table" "photos-users-dynamodb-table" {
  name           = "PhotosAppUsers"
  read_capacity  = 1
  write_capacity = 1
  hash_key       = "ID"

  attribute {
    name = "ID"
    type = "S"
  }

  attribute {
    name = "Username"
    type = "S"
  }

  global_secondary_index {
    name               = "Username-index"
    hash_key           = "Username"
    write_capacity     = 1
    read_capacity      = 1
    projection_type    = "ALL"
  }
}

resource "aws_dynamodb_table" "photos-photos-dynamodb-table" {
  name           = "PhotosAppPhotos"
  read_capacity  = 1
  write_capacity = 1
  hash_key       = "ID"

  attribute {
    name = "ID"
    type = "S"
  }

  attribute {
    name = "UserID"
    type = "S"
  }

  global_secondary_index {
    name               = "UserID-index"
    hash_key           = "UserID"
    write_capacity     = 1
    read_capacity      = 1
    projection_type    = "ALL"
  }
}

resource "aws_dynamodb_table" "photos-followers-dynamodb-table" {
  name           = "PhotosAppFollowers"
  read_capacity  = 1
  write_capacity = 1
  hash_key       = "UserID"
  range_key = "FollowerID"

  attribute {
    name = "UserID"
    type = "S"
  }

  attribute {
    name = "FollowerID"
    type = "S"
  }

  global_secondary_index {
    name               = "FollowerID-index"
    hash_key           = "FollowerID"
    write_capacity     = 1
    read_capacity      = 1
    projection_type    = "ALL"
  }
}

resource "aws_dynamodb_table" "photos-comments-dynamodb-table" {
  name           = "PhotosAppComments"
  read_capacity  = 1
  write_capacity = 1
  hash_key       = "PhotoID"
  range_key = "CreatedAt"

  attribute {
    name = "PhotoID"
    type = "S"
  }

  attribute {
    name = "CreatedAt"
    type = "S"
  }
}