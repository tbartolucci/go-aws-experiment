resource "aws_cognito_user_pool" "photos-app-pool" {
  name = "PhotosApp"
  mfa_configuration = "OFF"
  auto_verified_attributes = [ "email" ]
  sms_authentication_message = "Your authentication code is {####}. "
  sms_verification_message = "Your verification code is {####}. "
  password_policy {
    minimum_length    = 8
    require_lowercase = true
    require_numbers   = true
    require_symbols   = true
    require_uppercase = true
  }
}

resource "aws_cognito_user_pool_client" "photos-app-web" {
  name = "PhotosApp"

  user_pool_id = "${aws_cognito_user_pool.photos-app-pool.id}"
  explicit_auth_flows = ["ADMIN_NO_SRP_AUTH"]
}