resource "aws_lambda_permission" "allow_bucket" {
  statement_id  = "AllowExecutionFromS3Bucket"
  action        = "lambda:InvokeFunction"
  function_name = "${aws_lambda_function.photos-thumbnail-function.arn}"
  principal     = "s3.amazonaws.com"
  source_arn    = "${aws_s3_bucket.bitsbybit-pluralsight-photos.arn}"

}

resource "aws_lambda_function" "photos-thumbnail-function" {
  function_name     = "PhotosAppThumbnailFunction"
  role              = "${aws_iam_role.lambda_basic_s3_execution.arn}"
  handler           = "main"
  runtime           = "go1.x"
  s3_bucket         = "bitsbybit-pluralsight-photos-lambda"
  s3_key            = "main.zip"
  memory_size = "192"
  timeout = 30
  publish = "false"
}

resource "aws_s3_bucket_notification" "bucket_notification" {
  bucket = "${aws_s3_bucket.bitsbybit-pluralsight-photos.id}"

  lambda_function {
    lambda_function_arn = "${aws_lambda_function.photos-thumbnail-function.arn}"
    events              = ["s3:ObjectCreated:*"]
  }
}