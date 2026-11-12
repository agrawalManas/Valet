package constants

const DOBLayout = "2006-01-02"

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
	SITE_ID              = "site id"
	SUPERVISOR_ID        = "supervisor id"
	INVALID              = "Invalid"
	QUERY_PARAMS         = "Query Params"
	NOT_FOUND            = "Not Found"
)

// --------------File-Related-Constants
const (
	FILE_VISIBILITY_PRIVATE      = "PRIVATE"
	FILE_VISIBILITY_PUBLIC       = "PUBLIC"
	TOTAL_FILE_SIZE_LIMIT_EXCEED = "Total file upload limit is 50MB/request"
	FILE_LENGTH_EXCEED           = "Total no. of files in a request should be less than 10"
	REQUIRE_AT_LEAST_ONE_FILE    = "Please add at least one file to upload"
	INVALID_FORM_DATA            = "Invalid form data"
	FILE_KIND_IMAGE              = "IMAGE"
	FILE_KIND_FILE               = "FILE"
)

// ----------------Driver-Related-Constants
const (
	DRIVER_STATUS_AVAILABLE = "AVAILABLE"
	DRIVER_STATUS_ON_DUTY   = "ON_DUTY"
	DRIVER_STATUS_OFF_DUTY  = "OFF_DUTY"
	DRIVER_STATUS_SUSPENDED = "SUSPENDED"
	DRIVER_STATUS_INACTIVE  = "INACTIVE"
)

// ---------------KYC-Related-Constants
const (
	KYC_STATUS_PENDING   = "PENDING"
	KYC_STATUS_SUBMITTED = "SUBMITTED"
	KYC_STATUS_VERIFIED  = "VERIFIED"
	KYC_STATUS_REJECTED  = "REJECTED"
)
