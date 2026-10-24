package constants

const ROLE_CUSTOMER = "CUSTOMER"
const ROLE_DRIVER = "DRIVER"
const ROLE_SUPERVISOR = "SUPERVISOR"
const ROLE_ADMIN = "ADMIN"

const (
	CLAIM_USER_ID   = "user_id"
	CLAIM_USER_ROLE = "user_role"
	CLAIM_IS_ACTIVE = "is_active"
	CLAIM_EXP       = "exp"

	CTX_USER_ID   = "user_id"
	CTX_USER_ROLE = "user_role"
	CTX_AUTH_USER = "auth_user"
	CTX_IS_ACTIVE = "is_active"
)

const RESPONSE_STATUS_FAILED = "FAILED"
const RESPONSE_STATUS_SUCCESS = "SUCCESS"

const SOMETHING_WENT_WRONG = "Oops, Something went wrong"
const REQUEST_BODY_MISSING = "Provide a valid request body"
const INVALID_USER_ID = "Provide valid user id"
const USER_NOT_FOUND = "User not found"

const UNAUTHORIZED = "Unauthorized"
const FORBIDDEN = "Forbidden"
const TOKEN_EXPIRED = "TOKEN_EXPIRED"

const USER_EXISTS = "User with mobile number already exist"
const NEW_USER = "New user, please create your account"
const NO_USER_WITH_TOKEN = "Couldn't find the user associated with this token"
const USER_INACTIVE = "User inactive"
const NO_USER_ROLE = "User does not have a defined role"
const NO_PERMISSION = "You do not have permission(s) to perform this action"
const CANNOT_MODIFY_OWN_ACCESS = "Cannot modify own access"

const FAILED_TO_GENERATE_TOKEN = "Cannot get the token at the moment"
const NO_AUTH_HEADER = "No authorization header found"
const MALFORMED_TOKEN = "Malformed token"
const UNEXPECTED_SIGNING_METHOD = "Unexpected Signing Method"
const NO_TOKEN_CLAIMS_FOUND = "No token claims found"
const INVALID_TOKEN_CLAIM = "Invalid token claim"

const INVALID_MOBILE_NUMBER = "Please Provide a valid 10 digit mobile number"
const INVALID_OTP = "Invalid OTP"
