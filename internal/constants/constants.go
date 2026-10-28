package constants

// ---------Roles
const (
	ROLE_CUSTOMER   = "CUSTOMER"
	ROLE_DRIVER     = "DRIVER"
	ROLE_SUPERVISOR = "SUPERVISOR"
	ROLE_ADMIN      = "ADMIN"
)

// ----------Auth-Middleware
const (
	CLAIM_USER_ID   = "user_id"
	CLAIM_USER_ROLE = "user_role"
	CLAIM_IS_ACTIVE = "is_active"
	CLAIM_EXP       = "exp"
	CTX_AUTH_USER   = "auth_user"

	FAILED_TO_GENERATE_TOKEN  = "Cannot get the token at the moment"
	NO_AUTH_HEADER            = "No authorization header found"
	MALFORMED_TOKEN           = "Malformed token"
	UNEXPECTED_SIGNING_METHOD = "Unexpected Signing Method"
	NO_TOKEN_CLAIMS_FOUND     = "No token claims found"
	INVALID_TOKEN_CLAIM       = "Invalid token claim"
)

// -------------Response-Fields
const (
	RESPONSE_STATUS_FAILED  = "FAILED"
	RESPONSE_STATUS_SUCCESS = "SUCCESS"
)

// --------------Auth-Related-Errors
const (
	INVALID_MOBILE_NUMBER = "Please Provide a valid 10 digit mobile number"
	INVALID_OTP           = "Invalid OTP"
)

// ---------------User-Related-Errors
const (
	USER_EXISTS              = "User with mobile number already exist"
	NO_USER_WITH_TOKEN       = "Couldn't find the user associated with this token"
	USER_INACTIVE            = "User inactive"
	NO_USER_ROLE             = "User does not have a defined role"
	NO_PERMISSION            = "You do not have permission(s) to perform this action"
	CANNOT_MODIFY_OWN_ACCESS = "Cannot modify own access"
)

// -----------------Error-Message
const (
	SOMETHING_WENT_WRONG = "Oops, Something went wrong"
	INVALID_REQUEST_BODY = "Provide a valid request body"
	INVALID_USER_ID      = "Provide valid user id"
	USER_NOT_FOUND       = "User not found"
	UNAUTHORIZED         = "Unauthorized"
	FORBIDDEN            = "Forbidden"
	TOKEN_EXPIRED        = "Token Expired"
)
