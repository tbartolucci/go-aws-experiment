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

  schema {
    attribute_data_type      = "String"
    developer_only_attribute = false
    mutable                  = true
    name                     = "name"
    required                 = true

    string_attribute_constraints {
      min_length = 0
      max_length = 2048
    }
  }

  schema {
    attribute_data_type      = "String"
    developer_only_attribute = false
    mutable                  = true
    name                     = "email"
    required                 = true

    string_attribute_constraints {
      min_length = 0
      max_length = 2048
    }
  }

}

resource "aws_cognito_user_pool_client" "photos-app-web" {
  name = "PhotosApp"

  user_pool_id = "${aws_cognito_user_pool.photos-app-pool.id}"
  explicit_auth_flows = ["ADMIN_NO_SRP_AUTH"]
  read_attributes = ["given_name", "email_verified", "zoneinfo", "website", "preferred_username",
    "name", "locale", "phone_number", "family_name", "birthdate", "middle_name", "phone_number_verified",
    "profile", "picture", "address", "gender", "updated_at", "nickname", "email" ]
  write_attributes = ["given_name", "zoneinfo", "website", "preferred_username",
    "name", "locale", "phone_number", "family_name", "birthdate", "middle_name",
    "profile", "picture", "address", "gender", "updated_at", "nickname", "email" ]
}