//resource "aws_cognito_user_pool" "math-pool" {
//  name = "math-pool"
//  mfa_configuration = "OFF"
//  email_verification_message = "Your verification code is {####}. "
//  email_verification_subject = "Your Math Quiz verification code"
//  sms_authentication_message = "Your authentication code is {####}. "
//  sms_verification_message = "Your verification code is {####}. "
//  username_attributes = [ "email" ]
//  auto_verified_attributes = [ "email" ]
//  password_policy {
//    minimum_length    = 8
//    require_lowercase = true
//    require_numbers   = true
//    require_symbols   = true
//    require_uppercase = true
//  }
//  verification_message_template {
//    default_email_option = "CONFIRM_WITH_CODE"
//  }
//}