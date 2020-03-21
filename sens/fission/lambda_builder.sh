#---
fission fn update --name get-session --env go --src src/session/main.go --entrypoint GetSession
#---
fission fn update --name unregister-device --env go --src src/device/main.go --entrypoint UnregisterDevice
#---
fission fn update --name get-org-detail --env go --src src/auth/main.go --entrypoint GetOrgDetail
#---
fission fn update --name pair-device --env go --src src/device/main.go --entrypoint PairDevice
#---
fission fn update --name test-ds --env go --src src/test/main.go --entrypoint TestDatastore
#---
fission fn update --name list-sessions --env go --src src/session/main.go --entrypoint ListSessions
#---
fission fn update --name get-general-summary --env go --src src/session/main.go --entrypoint GetGeneralSummary
#---
fission fn update --name verify-otp --env go --src src/auth/main.go --entrypoint VerifyOtp
#---
fission fn update --name create-op --env go --src src/auth/main.go --entrypoint CreateOp
#---
fission fn update --name get-session-summary --env go --src src/session/main.go --entrypoint GetSessionSummary
#---
fission fn update --name request-otp --env go --src src/auth/main.go --entrypoint RequestOtp
#---
fission fn update --name create-auth --env go --src src/auth/main.go --entrypoint CreateAuth
#---
fission fn update --name create-org --env go --src src/auth/main.go --entrypoint CreateOrg
#---
fission fn update --name register-device --env go --src src/device/main.go --entrypoint RegisterDevice
#---
fission fn update --name unpair-device --env go --src src/device/main.go --entrypoint UnpairDevice
#---
fission fn update --name get-session-property --env go --src src/session/main.go --entrypoint GetSessionPropertyFunc
#---
fission fn update --name create-user --env go --src src/auth/main.go --entrypoint CreateUser
#---
fission fn update --name get-op-detail --env go --src src/auth/main.go --entrypoint GetOpDetail
#---
fission fn update --name get-user-detail --env go --src src/auth/main.go --entrypoint GetUserDetail
#---
fission fn update --name create-device --env go --src src/device/main.go --entrypoint CreateDevice
