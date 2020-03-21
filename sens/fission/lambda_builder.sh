#---
fission fn delete --name  unregister-device
fission fn create --name  unregister-device  --env go --src  src/device/main.go  --entrypoint  UnregisterDevice
fission route create --name unregister-device --method POST --url /api/devices/{deviceId}/unregister --function unregister-device
#---
fission fn delete --name  get-session-summary
fission fn create --name  get-session-summary  --env go --src  src/session/main.go  --entrypoint  GetSessionSummary
fission route create --name get-session-summary --method GET --url /api/sessions/{id}/summary --function get-session-summary
#---
fission fn delete --name  get-general-summary
fission fn create --name  get-general-summary  --env go --src  src/session/main.go  --entrypoint  GetGeneralSummary
fission route create --name get-general-summary --method GET --url /api/general-summary/get --function get-general-summary
#---
fission fn delete --name  create-org
fission fn create --name  create-org  --env go --src  src/auth/main.go  --entrypoint  CreateOrg
fission route create --name create-org --method POST --url /api/orgs/create --function create-org
#---
fission fn delete --name  get-op-detail
fission fn create --name  get-op-detail  --env go --src  src/auth/main.go  --entrypoint  GetOpDetail
fission route create --name get-op-detail --method POST --url /api/ops/{id}/detail --function get-op-detail
#---
fission fn delete --name  create-user
fission fn create --name  create-user  --env go --src  src/auth/main.go  --entrypoint  CreateUser
fission route create --name create-user --method POST --url /api/users/create --function create-user
#---
fission fn delete --name  create-device
fission fn create --name  create-device  --env go --src  src/device/main.go  --entrypoint  CreateDevice
fission route create --name create-device --method POST --url /api/devices/create --function create-device
#---
fission fn delete --name  create-auth
fission fn create --name  create-auth  --env go --src  src/auth/main.go  --entrypoint  CreateAuth
fission route create --name create-auth --method POST --url /api/auths/create --function create-auth
#---
fission fn delete --name  create-op
fission fn create --name  create-op  --env go --src  src/auth/main.go  --entrypoint  CreateOp
fission route create --name create-op --method POST --url /api/ops/create --function create-op
#---
fission fn delete --name  get-user-detail
fission fn create --name  get-user-detail  --env go --src  src/auth/main.go  --entrypoint  GetUserDetail
fission route create --name get-user-detail --method POST --url /api/users/{id}/detail --function get-user-detail
#---
fission fn delete --name  pair-device
fission fn create --name  pair-device  --env go --src  src/device/main.go  --entrypoint  PairDevice
fission route create --name pair-device --method POST --url /api/devices/{deviceId}/pair --function pair-device
#---
fission fn delete --name  test-ds
fission fn create --name  test-ds  --env go --src  src/test/main.go  --entrypoint  TestDatastore
fission route create --name test-ds --method POST --url /test --function test-ds
#---
fission fn delete --name  get-session
fission fn create --name  get-session  --env go --src  src/session/main.go  --entrypoint  GetSession
fission route create --name get-session --method GET --url /api/session/{id}/get --function get-session
#---
fission fn delete --name  request-otp
fission fn create --name  request-otp  --env go --src  src/auth/main.go  --entrypoint  RequestOtp
fission route create --name request-otp --method POST --url /api/otp/request --function request-otp
#---
fission fn delete --name  verify-otp
fission fn create --name  verify-otp  --env go --src  src/auth/main.go  --entrypoint  VerifyOtp
fission route create --name verify-otp --method POST --url /api/otp/verify --function verify-otp
#---
fission fn delete --name  unpair-device
fission fn create --name  unpair-device  --env go --src  src/device/main.go  --entrypoint  UnpairDevice
fission route create --name unpair-device --method POST --url /api/devices/{deviceId}/unpair --function unpair-device
#---
fission fn delete --name  list-sessions
fission fn create --name  list-sessions  --env go --src  src/session/main.go  --entrypoint  ListSessions
fission route create --name list-sessions --method GET --url /api/sessions/list --function list-sessions
#---
fission fn delete --name  get-session-property
fission fn create --name  get-session-property  --env go --src  src/session/main.go  --entrypoint  GetSessionPropertyFunc
fission route create --name get-session-property --method GET --url /api/session-property/{id}/get --function get-session-property
#---
fission fn delete --name  get-org-detail
fission fn create --name  get-org-detail  --env go --src  src/auth/main.go  --entrypoint  GetOrgDetail
fission route create --name get-org-detail --method POST --url /api/orgs/{id}/detail --function get-org-detail
#---
fission fn delete --name  register-device
fission fn create --name  register-device  --env go --src  src/device/main.go  --entrypoint  RegisterDevice
fission route create --name register-device --method POST --url /api/devices/{deviceId}/register --function register-device
