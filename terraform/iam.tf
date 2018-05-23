//resource "aws_iam_policy" "ecs_cloudwatch_logs" {
//  name        = "ecs-cloudwatch-logs"
//  path        = "/"
//  description = "ECS Cloudwatch Logging Policy"
//  policy      = <<EOF
//{
//    "Version": "2012-10-17",
//    "Statement": [
//        {
//            "Effect": "Allow",
//            "Action": [
//                "logs:CreateLogGroup",
//                "logs:CreateLogStream",
//                "logs:PutLogEvents",
//                "logs:DescribeLogStreams"
//            ],
//            "Resource": [
//                "arn:aws:logs:*:*:*"
//            ]
//        }
//    ]
//}
//EOF
//}
//
//resource "aws_iam_policy" "ecs_vpc_flow_log_policy" {
//  name        = "ecs-vpc-flow-logs"
//  path        = "/"
//  description = "ECS VPC Flow Log Role"
//  policy      = <<EOF
//{
//  "Version": "2012-10-17",
//  "Statement": [
//    {
//      "Action": [
//        "logs:CreateLogGroup",
//        "logs:CreateLogStream",
//        "logs:DescribeLogGroups",
//        "logs:DescribeLogStreams",
//        "logs:PutLogEvents"
//      ],
//      "Effect": "Allow",
//      "Resource": "*"
//    }
//  ]
//}
//EOF
//
//}
//
//resource "aws_iam_role" "ecs_vpc_flow_log_role" {
//  name = "ecs-vpc-flow-log-role"
//  assume_role_policy = <<EOF
//{
//  "Version": "2012-10-17",
//  "Statement": [
//    {
//      "Sid": "",
//      "Effect": "Allow",
//      "Principal": {
//        "Service": "vpc-flow-logs.amazonaws.com"
//      },
//      "Action": "sts:AssumeRole"
//    }
//  ]
//}
//EOF
//}
//
//resource "aws_iam_role_policy_attachment" "ecs-vpc-flow-log-attach" {
//  role       = "${aws_iam_role.ecs_vpc_flow_log_role.name}"
//  policy_arn = "${aws_iam_policy.ecs_vpc_flow_log_policy.arn}"
//}
