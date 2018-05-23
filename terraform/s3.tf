resource "aws_s3_bucket" "bitsbybit-pluralsight-photos" {
  bucket = "bitsbybit-pluralsight-photos"
  acl    = "private"
  policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": "*",
      "Action": "s3:GetObject",
      "Resource": "arn:aws:s3:::bitsbybit-pluralsight-photos/*"
    }
  ]
}
POLICY
}

resource "aws_s3_bucket" "bitsbybit-pluralsight-photos-lambda" {
  bucket = "bitsbybit-pluralsight-photos-lambda"
  acl    = "private"
}

