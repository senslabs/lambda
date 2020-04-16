fission env create --spec --name go --image dozeelabs/fission-go-env --builder dozeelabs/fission-go-builder --mincpu 40 --maxcpu 80 --minmemory 64 --maxmemory 128 --poolsize 3

fission function create --spec --name request-otp --env go --src "sens/ws/*" --entrypoint RequestOtp
fission route create --spec --method POST --url /api/otp/request --function request-otp --name request-otp

fission function create --spec --name verify-otp --env go --src "sens/ws/*" --entrypoint VerifyOtp
fission route create --spec --method POST --url /api/otp/verify --function verify-otp --name verify-otp

fission function create --spec --name verify-auth --env go --src "sens/ws/*" --entrypoint VerifyAuth
fission route create --spec --method POST --url /api/auth/verify --function verify-auth --name verify-auth

fission function create --spec --name add-org --env go --src "sens/ws/*" --entrypoint AddOrg
fission route create --spec --method POST --url /api/orgs/add --function add-org --name add-org

fission function create --spec --name create-org --env go --src "sens/ws/*" --entrypoint CreateOrg
fission route create --spec --method POST --url /api/orgs/create --function create-org --name create-org

fission function create --spec --name create-op --env go --src "sens/ws/*" --entrypoint CreateOp
fission route create --spec --method POST --url /api/ops/create --function create-op --name create-op

fission function create --spec --name add-op --env go --src "sens/ws/*" --entrypoint AddOp
fission route create --spec --method POST --url /api/ops/add --function add-op --name add-op

fission function create --spec --name create-user --env go --src "sens/ws/*" --entrypoint CreateUser
fission route create --spec --method POST --url /api/users/create --function create-user --name create-user

fission function create --spec --name add-user --env go --src "sens/ws/*" --entrypoint AddUser
fission route create --spec --method POST --url /api/users/add --function add-user --name add-user

fission function create --spec --name refresh-token --env go --src "sens/ws/*" --entrypoint RefreshToken
fission route create --spec --method GET --url /api/tokens/refresh --function refresh-token --name refresh-token

fission function create --spec --name change-role-to-op --env go --src "sens/ws/*" --entrypoint ChangeRoleToOp
fission route create --spec --method POST --url /api/roles/update/op --function change-role-to-op --name change-role-to-op

fission function create --spec --name change-role-to-auth --env go --src "sens/ws/*" --entrypoint ChangeRoleToAuth
fission route create --spec --method POST --url /api/roles/update/auth --function change-role-to-auth --name change-role-to-auth

fission function create --spec --name list-orgs --env go --src "sens/ws/*" --entrypoint ListOrgs
fission route create --spec --method GET --url /api/orgs/list --function list-orgs --name list-orgs

fission function create --spec --name list-ops --env go --src "sens/ws/*" --entrypoint ListOps
fission route create --spec --method GET --url /api/ops/list --function list-ops --name list-ops

fission function create --spec --name list-users --env go --src "sens/ws/*" --entrypoint ListUsers
fission route create --spec --method GET --url /api/users/list --function list-users --name list-users

fission function create --spec --name list-devices --env go --src "sens/ws/*" --entrypoint ListDevices
fission route create --spec --method GET --url /api/devices/list --function list-devices --name list-devices

fission function create --spec --name list-users-summary --env go --src "sens/ws/*" --entrypoint ListUserSummary
fission route create --spec --method GET --url /api/users/summary --function list-users-summary --name list-users-summary

fission function create --spec --name list-users-sleeps --env go --src "sens/ws/*" --entrypoint ListUserSleeps
fission route create --spec --method GET --url /api/users/sleeps --function list-users-sleeps --name list-users-sleeps

fission function create --spec --name list-users-meditations --env go --src "sens/ws/*" --entrypoint ListUserMeditations
fission route create --spec --method GET --url /api/users/meditations --function list-users-meditations --name list-users-meditations

fission function create --spec --name list-activities --env go --src "sens/ws/*" --entrypoint ListActivities
fission route create --spec --method GET --url /api/activities/list --function list-activities --name list-activities

fission function create --spec --name get-sleeps-summary --env go --src "sens/ws/*" --entrypoint GetSleepsSummary
fission route create --spec --method GET --url /api/sleeps/summary/get --function get-sleeps-summary --name get-sleeps-summary

fission function create --spec --name get-meditations-summary --env go --src "sens/ws/*" --entrypoint GetMeditationsSummary
fission route create --spec --method GET --url /api/meditations/summary/get --function get-meditations-summary --name get-meditations-summary

fission function create --spec --name list-alerts --env go --src "sens/ws/*" --entrypoint ListAlerts
fission route create --spec --method POST --url /api/alerts/list --function list-alerts --name list-alerts

fission function create --spec --name create-api-key --env go --src "sens/ws/*" --entrypoint CreateKey
fission route create --spec --method POST --url /api/keys/create --function create-api-key --name create-api-key

fission function create --spec --name list-api-keys --env go --src "sens/ws/*" --entrypoint ListKeys
fission route create --spec --method POST --url /api/keys/list --function list-api-keys --name list-api-keys

fission function create --spec --name create-device --env go --src "sens/ws/*" --entrypoint CreateDevice
fission route create --spec --method POST --url /api/devices/create --function create-device --name create-device

fission function create --spec --name register-device --env go --src "sens/ws/*" --entrypoint RegisterDevice
fission route create --spec --method POST --url /api/devices/register --function register-device --name register-device

fission function create --spec --name unregister-device --env go --src "sens/ws/*" --entrypoint UnregisterDevice
fission route create --spec --method POST --url /api/devices/unregister --function unregister-device --name unregister-device

fission function create --spec --name pair-device --env go --src "sens/ws/*" --entrypoint PairDevice
fission route create --spec --method POST --url /api/devices/pair --function pair-device --name pair-device

fission function create --spec --name unpair-device --env go --src "sens/ws/*" --entrypoint UnpairDevice
fission route create --spec --method POST --url /api/devices/unpair --function unpair-device --name unpair-device

fission function create --spec --name get-session --env go --src "sens/ws/*" --entrypoint GetSession
fission route create --spec --method POST --url /api/sessions/{sessionId}/get --function get-session --name get-session

