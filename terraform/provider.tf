#
# Provides IDs for resources created from this main Terraform repo
#
# Usage
#   - Copy the plugins to your Terraform directory by copying the
#     contents of (terraform repo)\plugins to the path where your
#     Terraform.exe is located, which you can get by `which terraform`
#     on linux, or `Get-Command terraform.exe` in Powershell.
#   - Run `terraform init` in your project's terraform directory to
#     install any provider plugins.
#

provider "aws" {
  region  = "us-east-1"
  access_key = "${var.aws_access_key}"
  secret_key = "${var.aws_secret_key}"
  version = "~> 1.10.0"
}

variable "region" {
  description = "The AWS region to create resources in."
  default = "us-east-1"
}

variable "acct" {
  type = "map"
  default = {
    account-id    = "123972417721"
    region        = "us-east-1"
    abbrev        = "personal"
    dnsabbrev     = "personal"
  }
}

variable "ec2_keypair_name" {
  default = "tom-desktop-key-pair"
}

data "aws_vpc" "default" {
  id = "vpc-07df587e"
}

data "aws_internet_gateway" "default" {
  internet_gateway_id = "igw-dc7700ba"
}

data "aws_route_table" "default" {
  route_table_id = "rtb-1aaee162"
}

data "aws_subnet" "us_east_1a" {
  id = "subnet-6b13a947"
}

data "aws_subnet" "us_east_1b" {
  id = "subnet-8f5f2dc7"
}

data "aws_subnet" "us_east_1c" {
  id = "subnet-7f01b325"
}

data "aws_subnet" "us_east_1d" {
  id = "subnet-612f3e04"
}

data "aws_subnet" "us_east_1e" {
  id = "subnet-baa89586"
}

data "aws_subnet" "us_east_1f" {
  id = "subnet-c0b939cc"
}

variable "tombartolucci_io_cert" {
  default = "arn:aws:acm:us-east-1:123972417721:certificate/91921ed3-71a8-4633-9177-9be151693092"
}

